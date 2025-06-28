package foundationdao

import (
	"context"
	"errors"
	"fmt"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"strconv"
	"strings"
)

type ProblemDao struct {
	db *gorm.DB
}

var singletonProblemDao = singleton.Singleton[ProblemDao]{}

func GetProblemDao() *ProblemDao {
	return singletonProblemDao.GetInstance(
		func() *ProblemDao {
			dao := &ProblemDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ProblemDao) GetProblemIdByKey(key string) (int, error) {
	if key == "" {
		return 0, metaerror.New("key is empty")
	}
	var id int
	err := d.db.
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("`key` = ?", key).
		Scan(&id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, metaerror.New("problem not found")
		}
		return 0, metaerror.Wrap(err, "find problem id by key failed")
	}
	if id == 0 {
		return 0, metaerror.New("problem not found")
	}
	return id, nil
}

func (d *ProblemDao) GetProblemList(
	ctx context.Context,
	oj string, title string, tags []int, private bool,
	userId int, hasAuth bool,
	page int,
	pageSize int,
) (
	[]*foundationmodel.Problem,
	int,
	error,
) {
	db := d.db.WithContext(ctx).Model(&foundationmodel.Problem{})
	if !hasAuth {
		if userId > 0 {
			// 拼接 WHERE (private IS NULL OR inserter = ? OR id IN problem_member OR id IN problem_member_auth)
			db = db.Where(
				"private IS NULL OR inserter = ? OR id IN (?) OR id IN (?)",
				userId,
				d.db.Model(&foundationmodel.ProblemMember{}).Select("id").Where("user_id = ?", userId),
				d.db.Model(&foundationmodel.ProblemMemberAuth{}).Select("id").Where("user_id = ?", userId),
			)
		} else {
			db = db.Where("private IS NULL")
		}
	} else if private {
		db = db.Where("private = 1")
	}
	if oj == "didaoj" {
		db = db.Where("origin_oj IS NULL")
	} else if oj != "" {
		db = db.Where("origin_oj = ?", oj)
	}
	if title != "" {
		db = db.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	if len(tags) > 0 {
		db = db.Joins("JOIN problem_tag pt ON pt.id = problem.id").
			Where("pt.tag_id IN ?", tags).
			Group("problem.id") // 防止重复行
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "count failed")
	}
	var list []*foundationmodel.Problem
	if err := db.
		Select("id", "title", "accept", "attempt").
		Order("sort ASC").
		Order("id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "find failed")
	}

	return list, int(total), nil
}

func (d *ProblemDao) InsertProblemLocal(
	ctx context.Context,
	problem *foundationmodel.Problem,
	problemLocal *foundationmodel.ProblemLocal,
) error {
	if problem == nil {
		return metaerror.New("problem is nil")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(problem).Error; err != nil {
				return metaerror.Wrap(err, "insert problem")
			}
			if problemLocal == nil {
				problemLocal = &foundationmodel.ProblemLocal{}
			}
			problemLocal.ProblemId = problem.Id
			if err := tx.Create(problemLocal).Error; err != nil {
				return metaerror.Wrap(err, "insert problem local")
			}
			problem.Key = strconv.Itoa(problemLocal.Id)
			if err := tx.Save(problem).Error; err != nil {
				return metaerror.Wrap(err, "update problem key")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *ProblemDao) InsertProblemRemote(
	ctx context.Context,
	problem *foundationmodel.Problem,
	problemRemote *foundationmodel.ProblemRemote,
) error {
	if problem == nil {
		return metaerror.New("problem is nil")
	}
	if problemRemote == nil {
		return metaerror.New("problemRemote is nil")
	}
	if problemRemote.OriginOj == "" || problemRemote.OriginId == "" {
		return metaerror.New("problemRemote originOj or originId is nil")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(problem).Error; err != nil {
				return metaerror.Wrap(err, "insert problem")
			}
			if problemRemote == nil {
				problemRemote = &foundationmodel.ProblemRemote{}
			}
			problemRemote.ProblemId = problem.Id
			if err := tx.Create(problemRemote).Error; err != nil {
				return metaerror.Wrap(err, "insert problem remote")
			}
			problem.Key = fmt.Sprintf("%s-%s", problemRemote.OriginOj, problemRemote.OriginId)
			if err := tx.Save(problem).Error; err != nil {
				return metaerror.Wrap(err, "update problem key")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}
