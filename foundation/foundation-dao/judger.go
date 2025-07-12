package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type JudgerDao struct {
	db *gorm.DB
}

var singletonJudgerDao = singleton.Singleton[JudgerDao]{}

func GetJudgerDao() *JudgerDao {
	return singletonJudgerDao.GetInstance(
		func() *JudgerDao {
			dao := &JudgerDao{}
			db := metamysql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model(&foundationmodel.Judger{})
			return dao
		},
	)
}

func (d *JudgerDao) UpdateJudger(ctx context.Context, judger *foundationmodel.Judger) error {
	err := d.db.WithContext(ctx).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}}, // 以 key 为唯一索引
				UpdateAll: true,                           // 冲突时更新所有字段
			},
		).
		Create(judger).Error

	if err != nil {
		return metaerror.Wrap(err, "failed to update judger")
	}
	return nil
}

func (d *JudgerDao) GetJudgers(ctx context.Context) ([]*foundationmodel.Judger, error) {
	var judgers []*foundationmodel.Judger
	err := d.db.WithContext(ctx).
		Where("hidden is null").
		Find(&judgers).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to get judgers")
	}
	return judgers, nil
}
