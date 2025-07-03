package foundationdao

import (
	"context"
	"errors"
	"fmt"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metatime "meta/meta-time"
	"meta/singleton"
)

type ProblemDailyDao struct {
	db *gorm.DB
}

var singletonProblemDailyDao = singleton.Singleton[ProblemDailyDao]{}

func GetProblemDailyDao() *ProblemDailyDao {
	return singletonProblemDailyDao.GetInstance(
		func() *ProblemDailyDao {
			dao := &ProblemDailyDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ProblemDailyDao) HasProblemDaily(ctx *gin.Context, dailyId string) (bool, error) {
	if dailyId == "" {
		return false, nil
	}
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ProblemDaily{}).
		Select("1").
		Where("`key` = ?", dailyId).
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

func (d *ProblemDailyDao) HasProblemDailyProblem(ctx *gin.Context, id int) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ProblemDaily{}).
		Select("1").
		Where("problem_id = ?", id).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, metaerror.Wrap(err, "check problem daily error")
	}
	return true, nil
}

func (d *ProblemDailyDao) GetProblemDaily(
	ctx *gin.Context,
	dailyId string,
	hasAuth bool,
) (*foundationmodel.ProblemDaily, error) {
	nowId := metatime.GetTimeNowBeijing().Format("2006-01-02")
	if !hasAuth {
		if dailyId > nowId {
			return nil, nil
		}
	}
	var record *foundationmodel.ProblemDaily
	err := d.db.WithContext(ctx).
		Select("problem_id,solution,code").
		Where("`key` = ?", dailyId).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "failed to find problem id by daily id: %s", dailyId)
	}
	return record, nil
}

func (d *ProblemDailyDao) GetProblemDailyEdit(ctx *gin.Context, dailyId string) (
	*foundationview.ProblemDailyEdit,
	error,
) {
	if dailyId == "" {
		return nil, metaerror.New("id is empty")
	}
	var record foundationview.ProblemDailyEdit
	err := d.db.WithContext(ctx).
		Select(
			"problem_daily.*, problem.`key` as problem_key, problem.title as problem_title, "+
				"inserter.username as inserter_username, inserter.nickname as inserter_nickname, "+
				"modifier.username as modifier_username, modifier.nickname as modifier_nickname",
		).
		Joins("JOIN problem ON problem.id = problem_daily.problem_id").
		Joins("JOIN user AS inserter ON inserter.id = problem_daily.inserter").
		Joins("JOIN user AS modifier ON modifier.id = problem_daily.modifier").
		Where("problem_daily.`key` = ?", dailyId).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "failed to find problem daily edit by id: %s", dailyId)
	}
	return &record, nil
}

func (d *ProblemDailyDao) GetDailyList(
	ctx context.Context,
	hasAuth bool,
	startDate *string,
	endDate *string,
	problemId string,
	page int,
	pageSize int,
) ([]*foundationview.ProblemDailyList, int, error) {
	db := d.db.WithContext(ctx).Table("problem_daily as pd")
	nowId := metatime.GetTimeNowBeijing().Format("2006-01-02")
	if startDate != nil && *startDate != "" {
		db = db.Where("pd.`key` >= ?", *startDate)
	}
	if hasAuth {
		if endDate != nil && *endDate != "" {
			db = db.Where("pd.`key` <= ?", *endDate)
		}
	} else {
		if endDate != nil && *endDate != "" {
			if *endDate < nowId {
				db = db.Where("pd.`key` <= ?", *endDate)
			} else {
				db = db.Where("pd.`key` <= ?", nowId)
			}
		} else {
			db = db.Where("pd.`key` <= ?", nowId)
		}
	}
	if problemId != "" {
		db = db.Where("pd.problem_id = ?", problemId)
	}
	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count problem daily: %w", err)
	}
	offset := (page - 1) * pageSize
	var list []*foundationview.ProblemDailyList
	err := db.Select(
		"pd.`key`",
		"pd.problem_id",
		"p.title AS title",
		"p.`key` AS problem_key",
		"p.accept AS accept",
		"p.attempt AS attempt",
	).
		Joins("LEFT JOIN problem AS p ON pd.problem_id = p.id").
		Order("`key` DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query problem daily list: %w", err)
	}
	return list, int(totalCount), nil
}

func (d *ProblemDailyDao) GetDailyRecently(ctx *gin.Context) ([]*foundationview.ProblemDailyList, error) {
	// 获取今天日期
	today := metatime.GetTimeNowBeijing().Format("2006-01-02")
	var result []*foundationview.ProblemDailyList
	err := d.db.WithContext(ctx).
		Table("problem_daily AS pd").
		Select(
			"`pd`.`key` AS `key`",
			"p.title AS title",
			"pd.problem_id AS problem_id",
			"p.`key` AS problem_key",
		).
		Joins("LEFT JOIN problem AS p ON pd.problem_id = p.id").
		Where("pd.`key` <= ?", today).
		Order("pd.`key` DESC").
		Limit(7).
		Scan(&result).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to query problem daily")
	}
	return result, nil
}

func (d *ProblemDailyDao) InsertProblemDaily(
	ctx context.Context,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	if problemDaily == nil {
		return metaerror.New("problemDaily is nil")
	}
	db := d.db.WithContext(ctx).Model(problemDaily)
	if err := db.Create(problemDaily).Error; err != nil {
		return metaerror.Wrap(err, "insert problemDaily")
	}
	return nil
}

func (d *ProblemDailyDao) UpdateProblemDaily(
	ctx context.Context,
	key string,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	if problemDaily == nil {
		return metaerror.New("daily is nil")
	}
	db := d.db.WithContext(ctx).Model(problemDaily)
	if err := db.Where("`key` = ?", key).Updates(problemDaily).Error; err != nil {
		return metaerror.Wrap(err, "update problemDaily")
	}
	if db.RowsAffected == 0 {
		return metaerror.New("problemDaily not found")
	}
	return nil
}
