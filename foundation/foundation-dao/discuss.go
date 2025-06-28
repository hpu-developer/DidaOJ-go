package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type DiscussDao struct {
	db *gorm.DB
}

var singletonDiscussDao = singleton.Singleton[DiscussDao]{}

func GetDiscussDao() *DiscussDao {
	return singletonDiscussDao.GetInstance(
		func() *DiscussDao {
			dao := &DiscussDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
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
