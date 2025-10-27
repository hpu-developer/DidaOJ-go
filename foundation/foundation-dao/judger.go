package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JudgerDao struct {
	db *gorm.DB
}

var singletonJudgerDao = singleton.Singleton[JudgerDao]{}

func GetJudgerDao() *JudgerDao {
	return singletonJudgerDao.GetInstance(
		func() *JudgerDao {
			dao := &JudgerDao{}
			db := metapostgresql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model(&foundationmodel.Judger{})
			return dao
		},
	)
}

func (d *JudgerDao) IsEnableJudge(ctx context.Context, key string) (bool, error) {
	var exists bool
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Judger{}).
		Select("1").
		Where("key = ? AND enable = TRUE", key).
		Limit(1).
		Scan(&exists).Error
	if err != nil {
		return false, metaerror.Wrap(err, "failed to check judger enable state")
	}
	return exists, nil
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
