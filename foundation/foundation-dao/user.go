package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type UserDao struct {
	db *gorm.DB
}

var singletonUserDao = singleton.Singleton[UserDao]{}

func GetUserDao() *UserDao {
	return singletonUserDao.GetInstance(
		func() *UserDao {
			dao := &UserDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *UserDao) InsertUser(
	ctx context.Context,
	user *foundationmodel.User,
) error {
	if user == nil {
		return metaerror.New("user is nil")
	}
	db := d.db.WithContext(ctx).Model(user)
	if err := db.Create(user).Error; err != nil {
		return metaerror.Wrap(err, "insert user")
	}
	return nil
}
