package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
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

func (d *UserDao) GetUserLogin(ctx context.Context, id int) (*foundationview.UserLogin, error) {
	var user foundationview.UserLogin
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname, password").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user login info")
	}
	return &user, nil
}

func (d *UserDao) GetInfoByUsername(ctx context.Context, username string) (*foundationview.UserInfo, error) {
	if username == "" {
		return nil, metaerror.New("username is empty")
	}
	var userInfo foundationview.UserInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select(
			`id, username, nickname, real_name, 
email, gender, number, slogan, organization, qq, 
vjudge_id, github, codeforces, 
check_in_count, insert_time, modify_time, accept, attempt`,
		).
		Where("username = ?", username).
		First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user info by username")
	}
	return &userInfo, nil
}

func (d *UserDao) GetUserAccountInfos(ctx context.Context, ids []int) ([]*foundationview.UserAccountInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var userAccountInfos []*foundationview.UserAccountInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname").
		Where("id IN ?", ids).
		Find(&userAccountInfos).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get user account infos")
	}
	return userAccountInfos, nil

}

func (d *UserDao) InsertUser(ctx context.Context, user *foundationmodel.User) error {
	if user == nil {
		return metaerror.New("user is nil")
	}
	db := d.db.WithContext(ctx).Model(user)
	if err := db.Create(user).Error; err != nil {
		return metaerror.Wrap(err, "insert user")
	}
	return nil
}
