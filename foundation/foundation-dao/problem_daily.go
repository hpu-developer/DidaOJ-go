package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
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
