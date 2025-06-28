package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type JudgeJobDao struct {
	db *gorm.DB
}

var singletonJudgeJobDao = singleton.Singleton[JudgeJobDao]{}

func GetJudgeJobDao() *JudgeJobDao {
	return singletonJudgeJobDao.GetInstance(
		func() *JudgeJobDao {
			dao := &JudgeJobDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *JudgeJobDao) InsertJudgeJob(
	ctx context.Context,
	judgeJob *foundationmodel.JudgeJob,
) error {
	if judgeJob == nil {
		return metaerror.New("judgeJob is nil")
	}
	if err := d.db.WithContext(ctx).Create(judgeJob).Error; err != nil {
		return metaerror.Wrap(err, "insert judgeJob")
	}
	return nil
}
