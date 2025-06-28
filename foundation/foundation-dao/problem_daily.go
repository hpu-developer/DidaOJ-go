package foundationdao

import (
	"context"
	"errors"
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
		Where("key = ?", dailyId).
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
		Where("id = ?", dailyId).
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
			"problem_daily.*, problem.key as problem_key, problem.title as problem_title, "+
				"inserter.username as inserter_username, inserter.nickname as inserter_nickname, "+
				"modifier.username as modifier_username, modifier.nickname as modifier_nickname",
		).
		Joins("JOIN problem ON problem.id = problem_daily.problem_id").
		Joins("JOIN user AS inserter ON inserter.id = problem_daily.inserter").
		Joins("JOIN user AS modifier ON modifier.id = problem_daily.modifier").
		Where("problem_daily.key = ?", dailyId).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "failed to find problem daily edit by id: %s", dailyId)
	}
	return &record, nil
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
