package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DiscussDao struct {
	db *gorm.DB
}

var singletonDiscussDao = singleton.Singleton[DiscussDao]{}

func GetDiscussDao() *DiscussDao {
	return singletonDiscussDao.GetInstance(
		func() *DiscussDao {
			dao := &DiscussDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *DiscussDao) IsDiscussBannedOrNotExist(ctx context.Context, id int) (bool, error) {
	var banned bool
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.Discuss{}).
		Where("id = ?", id).
		Pluck("banned", &banned).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, metaerror.Wrap(err, "find discuss error")
	}
	return banned, nil
}

func (d *DiscussDao) GetDiscussDetail(ctx context.Context, id int) (*foundationview.DiscussDetail, error) {
	var discuss foundationview.DiscussDetail
	err := d.db.WithContext(ctx).
		Table("discuss AS d").
		Select(
			`
			d.id, d.title, d.content, d.view_count, d.problem_id, d.contest_id, d.judge_id,
			d.inserter, d.modifier, d.insert_time, d.modify_time, d.update_time, d.banned,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname, u1.email AS inserter_email,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname,
			p.key AS problem_key
		`,
		).
		Joins("LEFT JOIN \"user\" AS u1 ON d.inserter = u1.id").
		Joins("LEFT JOIN \"user\" AS u2 ON d.modifier = u2.id").
		Joins("LEFT JOIN problem AS p ON d.problem_id = p.id").
		Where("d.id = ?", id).
		Scan(&discuss).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, metaerror.New("discuss not found")
		}
		return nil, metaerror.Wrap(err, "find discuss error")
	}
	return &discuss, nil
}

func (d *DiscussDao) GetInserter(ctx context.Context, id int) (int, error) {
	var inserter int
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.Discuss{}).
		Where("id = ?", id).
		Pluck("inserter", &inserter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, metaerror.New("discuss not found")
		}
		return 0, metaerror.Wrap(err, "get discuss author id failed")
	}
	return inserter, nil
}

func (d *DiscussDao) GetContent(ctx context.Context, id int) (*string, error) {
	var content string
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.Discuss{}).
		Where("id = ?", id).
		Pluck("content", &content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discuss content error")
	}
	if content == "" {
		return nil, nil
	}
	return &content, nil
}

func (d *DiscussDao) GetDiscussList(
	ctx context.Context,
	onlyProblem bool,
	contestId int,
	problemId int,
	title string,
	userId int,
	page int,
	pageSize int,
) ([]*foundationview.DiscussList, int, error) {
	db := d.db.WithContext(ctx).Model(&foundationmodel.Discuss{})
	// contest_id 条件
	if contestId > 0 {
		db = db.Where("contest_id = ?", contestId)
	} else {
		// Mongo 的 "$exists: false" 等价于 SQL 的 "IS NULL"
		db = db.Where("contest_id IS NULL")
	}

	// problem_id 条件
	if problemId > 0 {
		db = db.Where("discuss.problem_id = ?", problemId)
	} else if onlyProblem {
		// Mongo 的 "$exists: true" 等价于 SQL 的 "IS NOT NULL"
		db = db.Where("discuss.problem_id IS NOT NULL")
	}

	// title 模糊匹配，等价于 Mongo 的 "$regex" + "$options: i"
	if title != "" {
		db = db.Where("discuss.title ILIKE ?", "%"+title+"%")
	}

	// inserter 条件
	if userId > 0 {
		db = db.Where("discuss.inserter = ?", userId)
	}

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count records, page: %d", page)
	}
	// 分页查询
	var list []*foundationview.DiscussList
	selectFields := []string{
		"discuss.id",
		"discuss.title",
		"discuss.insert_time",
		"discuss.modify_time",
		"discuss.update_time",
		"discuss.inserter",
		"discuss.modifier",
		"discuss.view_count",
		"discuss.contest_id",
		"discuss.problem_id",

		"ui.username AS inserter_username",
		"ui.nickname AS inserter_nickname",
		"ui.email AS inserter_email",
		"um.username AS modifier_username",
		"um.nickname AS modifier_nickname",
	}

	if contestId > 0 {
		selectFields = append(selectFields, "cp.index AS contest_problem_index")
	} else {
		selectFields = append(selectFields, "p.key AS problem_key")
	}

	// 构造 DB 查询
	db = db.Select(selectFields).
		Joins("LEFT JOIN \"user\" AS ui ON ui.id = discuss.inserter").
		Joins("LEFT JOIN \"user\" AS um ON um.id = discuss.modifier")

	if contestId > 0 {
		db = db.Joins("LEFT JOIN contest_problem AS cp ON cp.problem_id = discuss.problem_id AND cp.id = discuss.contest_id")
	} else {
		db = db.Joins("LEFT JOIN problem AS p ON p.id = discuss.problem_id")
	}

	err := db.
		Order("discuss.update_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to find records, page: %d", page)
	}
	return list, int(total), nil
}

func (d *DiscussDao) InsertDiscuss(ctx context.Context, discuss *foundationmodel.Discuss) error {
	if discuss == nil {
		return metaerror.New("discuss is nil")
	}
	db := d.db.WithContext(ctx).Model(&foundationmodel.Discuss{})
	if err := db.Create(discuss).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return metaerror.New("discuss already exists")
		}
		return metaerror.Wrap(err, "insert discuss failed")
	}
	if discuss.Id == 0 {
		return metaerror.New("discuss id is zero after insert")
	}
	return nil
}

func (d *DiscussDao) UpdateContent(ctx context.Context, id int, content string) error {
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.Discuss{}).
		Where("id = ?", id).
		Update("content", content).Error; err != nil {
		return metaerror.Wrap(err, "update discuss content failed")
	}
	return nil
}

func (d *DiscussDao) PostEdit(ctx *gin.Context, discussId int, discuss *foundationmodel.Discuss) error {
	updateData := map[string]interface{}{
		"title":       discuss.Title,
		"content":     discuss.Content,
		"modify_time": discuss.ModifyTime,
		"update_time": discuss.UpdateTime,
	}
	if discuss.ProblemId != nil {
		updateData["problem_id"] = discuss.ProblemId
	} else {
		updateData["problem_id"] = nil // GORM 会将其设置为 NULL
	}
	if discuss.ContestId != nil {
		updateData["contest_id"] = discuss.ContestId
	} else {
		updateData["contest_id"] = nil // GORM 会将其设置为 NULL
	}
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Discuss{}).
		Where("id = ?", discussId).
		Updates(updateData).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to save discuss")
	}
	return nil
}
