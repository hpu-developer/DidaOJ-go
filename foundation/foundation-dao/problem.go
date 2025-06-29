package foundationdao

import (
	"context"
	"errors"
	"fmt"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metatime "meta/meta-time"
	"meta/singleton"
	"strconv"
	"strings"
)

type ProblemDao struct {
	db      *gorm.DB
	dbLocal *gorm.DB
}

var singletonProblemDao = singleton.Singleton[ProblemDao]{}

func GetProblemDao() *ProblemDao {
	return singletonProblemDao.GetInstance(
		func() *ProblemDao {
			dao := &ProblemDao{}
			db := metamysql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model(&foundationmodel.Problem{})
			dao.dbLocal = db.Model(&foundationmodel.ProblemLocal{})
			return dao
		},
	)
}

func (d *ProblemDao) GetProblemList(
	ctx context.Context,
	oj string, title string, tags []int, private bool,
	userId int, hasAuth bool,
	page int,
	pageSize int,
) (
	[]*foundationview.ProblemViewList,
	int,
	error,
) {
	db := d.db.WithContext(ctx).Model(&foundationmodel.Problem{})
	if !hasAuth {
		if userId > 0 {
			db = db.Where(
				"private = 0 OR inserter = ? OR id IN (?) OR id IN (?)",
				userId,
				d.db.Model(&foundationmodel.ProblemMember{}).Select("id").Where("user_id = ?", userId),
				d.db.Model(&foundationmodel.ProblemMemberAuth{}).Select("id").Where("user_id = ?", userId),
			)
		} else {
			db = db.Where("private = 0")
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
	var list []*foundationview.ProblemViewList
	if err := db.
		Select("id", "title", "accept", "attempt").
		Order("id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "find failed")
	}
	return list, int(total), nil
}

func (d *ProblemDao) GetProblemView(
	ctx context.Context, id string, userId int, hasAuth bool,
) (*foundationview.Problem, error) {
	db := d.db.WithContext(ctx).Table("problems AS p").
		Select(
			`
			p.*,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname,
			r.origin_oj, r.origin_id, r.origin_url, r.origin_author
		`,
		).
		Joins(`LEFT JOIN users u1 ON u1.id = p.inserter`).
		Joins(`LEFT JOIN users u2 ON u2.id = p.modifier`).
		Joins(`LEFT JOIN problem_remote r ON r.problem_id = p.id`).
		Where("p.id = ?", id)

	if !hasAuth {
		if userId > 0 {
			db = db.Where(
				`
				p.private = 0 OR
				p.inserter = ? OR
				EXISTS (
					SELECT 1 FROM problem_member pm WHERE pm.problem_id = p.id AND pm.user_id = ?
				) OR
				EXISTS (
					SELECT 1 FROM problem_auth_member pam WHERE pam.problem_id = p.id AND pam.user_id = ?
				)
			`, userId, userId, userId,
			)
		} else {
			db = db.Where("p.private = 0")
		}
	}

	var problem foundationview.Problem
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) CheckProblemEditAuth(ctx context.Context, problemId string, userId int) (
	bool,
	error,
) {
	var exists int
	err := d.db.WithContext(ctx).Raw(
		`SELECT 1
		FROM problem p
		LEFT JOIN problem_member_auth pa ON p.id = pa.id AND pa.user_id = ?
		WHERE p.id = ? AND (p.creator_id = ? OR pa.user_id IS NOT NULL)
		LIMIT 1
	`, userId, problemId, userId,
	).Scan(&exists).Error

	if err != nil {
		return false, metaerror.Wrap(err, "check edit permission failed")
	}
	return exists == 1, nil
}

func (d *ProblemDao) CheckProblemSubmitAuth(ctx context.Context, problemId int, userId int) (
	bool,
	error,
) {
	var exists int
	err := d.db.WithContext(ctx).Raw(
		`
		SELECT 1
		FROM problem p
		LEFT JOIN problem_member m ON p.id = m.id AND m.user_id = ?
		LEFT JOIN problem_member_auth a ON p.id = a.id AND a.user_id = ?
		WHERE p.id = ?
		  AND (
		    p.private = false
		    OR p.creator_id = ?
		    OR m.user_id IS NOT NULL
		    OR a.user_id IS NOT NULL
		  )
		LIMIT 1;
	`, userId, userId, problemId, userId,
	).Scan(&exists).Error
	if err != nil {
		return false, metaerror.Wrap(err, "check submit permission failed")
	}
	return exists == 1, nil
}

func (d *ProblemDao) GetProblemViewAuth(ctx context.Context, id string) (*foundationview.ProblemViewAuth, error) {
	var problem foundationview.ProblemViewAuth
	tx := d.db.WithContext(ctx)
	if err := tx.
		Select("id", "creator_id", "private").
		Where("id = ?", id).
		Take(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	if err := tx.
		Model(&foundationmodel.ProblemMember{}).
		Where("id = ?", id).
		Pluck("user_id", &problem.Members).Error; err != nil {
		return nil, metaerror.Wrap(err, "find problem members error")
	}
	if err := tx.
		Model(&foundationmodel.ProblemMemberAuth{}).
		Where("id = ?", id).
		Pluck("user_id", &problem.AuthMembers).Error; err != nil {
		return nil, metaerror.Wrap(err, "find problem auth members error")
	}
	return &problem, nil
}

func (d *ProblemDao) HasProblem(ctx context.Context, id int) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("1").
		Where("id = ?", id).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *ProblemDao) HasProblemByKey(ctx context.Context, key string) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("1").
		Where("key = ?", key).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *ProblemDao) HasProblemTitle(ctx context.Context, title string) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("1").
		Where("title = ?", title).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, metaerror.Wrap(err, "check problem title error")
	}
	return true, nil
}

func (d *ProblemDao) GetProblemIdByKey(ctx context.Context, key string) (int, error) {
	var id int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("`key` = ?", key).
		Pluck("id", &id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, metaerror.New("problem not found")
		}
		return 0, metaerror.Wrap(err, "find problem id by key failed")
	}
	return id, nil
}

func (d *ProblemDao) GetProblemIdsByKey(ctx context.Context, problemKeys []string) ([]int, error) {
	if len(problemKeys) == 0 {
		return nil, nil
	}
	var ids []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("`key` IN ?", problemKeys).
		Scan(&ids).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "find problem ids by keys failed")
	}
	return ids, nil
}

func (d *ProblemDao) GetProblemTitle(ctx context.Context, id string) (*string, error) {
	var problem foundationmodel.Problem
	err := d.db.WithContext(ctx).
		Select("title").
		Where("id = ?", id).
		Take(&problem).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &problem.Title, err
}

func (d *ProblemDao) GetProblemTitles(
	ctx context.Context,
	userId int,
	hasAuth bool,
	ids []int,
) ([]*foundationview.ProblemViewTitle, error) {
	query := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id", "title").
		Where("id IN ?", ids)
	if !hasAuth {
		if userId > 0 {
			query = query.Where(
				`
				private = 0 OR
				inserter = ? OR
				EXISTS (
					SELECT 1 FROM problem_member pm
					WHERE pm.problem_id = problems.id AND pm.user_id = ?
				) OR
				EXISTS (
					SELECT 1 FROM problem_auth_member pam
					WHERE pam.problem_id = problems.id AND pam.user_id = ?
				)
			`, userId, userId, userId,
			)
		} else {
			query = query.Where("private = 0")
		}
	}
	var titles []*foundationview.ProblemViewTitle
	err := query.Find(&titles).Error
	return titles, err
}

func (d *ProblemDao) GetProblemViewJudgeData(ctx context.Context, id string) (*foundationview.ProblemJudgeData, error) {
	db := d.db.WithContext(ctx).Table("problems AS p").
		Select(
			`
			p.id, p.key, p.title, p.judge_type,p.inserter, p.insert_time, p.modifier, p.modify_time,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname,
			r.judge_md5
		`,
		).
		Joins(`LEFT JOIN users u1 ON u1.id = p.inserter`).
		Joins(`LEFT JOIN users u2 ON u2.id = p.modifier`).
		Joins(`LEFT JOIN problem_local r ON r.problem_id = p.id`).
		Where("p.id = ?", id)
	var problem foundationview.ProblemJudgeData
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewApproveJudge(
	ctx context.Context,
	id int,
) (*foundationview.ProblemViewApproveJudge, error) {
	db := d.db.WithContext(ctx).
		Model(&foundationview.ProblemViewApproveJudge{}).
		Select("id", "origin_oj", "origin_id").
		Where("problem_id = ?", id)
	var problem foundationview.ProblemViewApproveJudge
	err := db.First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem approve judge error")
	}
	return nil, nil
}

func (d *ProblemDao) GetProblemJudgeMd5(ctx context.Context, id string) (*string, error) {
	var result struct {
		JudgeMd5 *string
	}
	err := d.dbLocal.WithContext(ctx).
		Select("judge_md5").
		Where("id = ?", id).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result.JudgeMd5, err
}

func (d *ProblemDao) GetProblemDescription(ctx context.Context, id string) (*string, error) {
	var result struct {
		Description string `gorm:"column:description"`
	}
	err := d.dbLocal.WithContext(ctx).
		Select("description").
		Where("id = ?", id).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &result.Description, err
}

func (d *ProblemDao) GetProblemListTitle(ctx context.Context, ids []string) (
	[]*foundationview.ProblemViewTitle,
	error,
) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*foundationview.ProblemViewTitle
	err := d.db.WithContext(ctx).
		Select("id", "title").
		Where("id IN ?", ids).
		Find(&list).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get problem list title error")
	}
	return list, nil
}

func (d *ProblemDao) FilterValidProblemIds(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var validIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("id IN ?", ids).
		Pluck("id", &validIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "find problems error")
	}
	return validIds, nil
}

func (d *ProblemDao) SelectProblemViewList(ctx context.Context, ids []int, needAttempt bool) (
	[]*foundationview.ProblemViewList,
	error,
) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*foundationview.ProblemViewList
	fields := []string{"id", "key", "title"}
	if needAttempt {
		fields = append(fields, "accept", "attempt")
	}
	err := d.db.WithContext(ctx).
		Select(fields).
		Where("id IN ?", ids).
		Find(&list).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get problems error")
	}
	return list, nil
}

func (d *ProblemDao) UpdateProblem(
	ctx context.Context,
	problemId string,
	problem *foundationmodel.Problem,
	tags []string,
) error {
	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			var tagIds []int
			for _, tagName := range tags {
				id, err := GetTagDao().InsertTagWithDb(tx, tagName)
				if err != nil {
					return err
				}
				tagIds = append(tagIds, id)
			}
			updateData := map[string]interface{}{
				"title":        problem.Title,
				"description":  problem.Description,
				"time_limit":   problem.TimeLimit,
				"memory_limit": problem.MemoryLimit,
				"source":       problem.Source,
				"modify_time":  problem.ModifyTime,
				"private":      problem.Private,
			}
			err := tx.Model(&foundationmodel.Problem{}).
				Where("id = ?", problemId).
				Updates(updateData).Error
			if err != nil {
				return err
			}
			err = GetProblemTagDao().UpdateProblemTagsByDb(tx, problem.Id, tagIds)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed to update problem")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemDescription(
	ctx context.Context,
	id int,
	description string,
) error {
	nowTime := metatime.GetTimeNow()
	err := d.db.WithContext(ctx).
		Where("id = ?", id).
		Update("description", description).
		Update("modify_time", nowTime).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return metaerror.New("problem not found")
		}
		return metaerror.Wrap(err, "failed to update problem description")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemJudgeInfo(
	ctx context.Context,
	id string,
	judgeType foundationjudge.JudgeType,
	md5 string,
) error {
	nowTime := metatime.GetTimeNow()
	err := d.db.Transaction(
		func(tx *gorm.DB) error {
			err := tx.Model(&foundationmodel.Problem{}).
				Where("id = ?", id).
				Update("judge_type", judgeType).
				Update("modify_time", nowTime).
				Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("problem not found")
				}
				return metaerror.Wrap(err, "failed to update problem description")
			}
			err = tx.Model(&foundationmodel.ProblemLocal{}).
				Where("id = ?", id).
				Update("judge_md5", md5).
				Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("problem local not found")
				}
				return metaerror.Wrap(err, "failed to update problem local judge md5")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed to update problem judge info")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemCrawl(
	ctx context.Context,
	id string,
	problem *foundationmodel.Problem,
	problemRemote *foundationmodel.ProblemRemote,
) error {
	err := d.db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(&foundationmodel.Problem{}).
				Where("id = ?", id).
				Save(problem).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("problem not found")
				}
				return metaerror.Wrap(err, "failed to update problem crawl")
			}
			if err := tx.Model(&foundationmodel.ProblemRemote{}).
				Where("problem_id = ?", id).
				Save(problemRemote).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("problem remote not found")
				}
				return metaerror.Wrap(err, "failed to update problem remote crawl")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed to update problem crawl")
	}
	return nil
}

func (d *ProblemDao) InsertProblemLocal(
	ctx context.Context,
	problem *foundationmodel.Problem,
	problemLocal *foundationmodel.ProblemLocal,
	tags []string,
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
