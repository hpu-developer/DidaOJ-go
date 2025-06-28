package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type TagDao struct {
	collection *mongo.Collection
	db         *gorm.DB
}

var singletonTagDao = singleton.Singleton[TagDao]{}

func GetTagDao() *TagDao {
	return singletonTagDao.GetInstance(
		func() *TagDao {
			dao := &TagDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *TagDao) GetClient() *gorm.DB {
	return metamysql.GetSubsystem().GetClient("didaoj")
}

func (d *TagDao) InsertTag(
	ctx context.Context,
	name string,
) error {
	if name == "" {
		return metaerror.New("problem is nil")
	}
	db := d.GetClient().WithContext(ctx).Model(&foundationmodel.Tag{})
	if err := db.Create(
		&foundationmodel.Tag{
			Name: name,
		},
	).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}
