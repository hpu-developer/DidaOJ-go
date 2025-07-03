package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type DiscussTagDao struct {
	db *gorm.DB
}

var singletonDiscussTagDao = singleton.Singleton[DiscussTagDao]{}

func GetDiscussTagDao() *DiscussTagDao {
	return singletonDiscussTagDao.GetInstance(
		func() *DiscussTagDao {
			dao := &DiscussTagDao{}
			db := metamysql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model((*foundationmodel.DiscussTag)(nil))
			return dao
		},
	)
}

func (d *DiscussTagDao) GetDiscussTags(ctx context.Context, discussIds int) ([]*foundationmodel.DiscussTag, error) {
	var tags []*foundationmodel.DiscussTag
	if err := d.db.WithContext(ctx).
		Where("id = ?", discussIds).
		Find(&tags).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to find discuss tags")
	}
	return tags, nil
}
