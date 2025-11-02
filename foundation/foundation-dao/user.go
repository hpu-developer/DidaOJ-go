package foundationdao

import (
	"context"
	"errors"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	"foundation/foundation-request"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

var singletonUserDao = singleton.Singleton[UserDao]{}

func GetUserDao() *UserDao {
	return singletonUserDao.GetInstance(
		func() *UserDao {
			dao := &UserDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
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

func (d *UserDao) GetUserLoginByUsername(ctx context.Context, username string) (*foundationview.UserLogin, error) {
	var user foundationview.UserLogin
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname, password").
		Where("LOWER(username) = LOWER(?)", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user login info by username")
	}
	return &user, nil
}

func (d *UserDao) GetUserPassword(ctx context.Context, userId int) (string, error) {
	var password string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("password").
		Where("id = ?", userId).
		Pluck("password", &password).Error
	if err != nil {
		return "", metaerror.Wrap(err, "get user password")
	}
	return password, nil
}

func (d *UserDao) GetModifyInfo(ctx context.Context, userId int) (*foundationview.UserModifyInfo, error) {
	var userInfo foundationview.UserModifyInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select(
			`id, username, nickname, real_name, 
email, gender, number, slogan, organization, qq, 
vjudge_id, github, codeforces`,
		).
		Where("id = ?", userId).
		First(&userInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, metaerror.Wrap(err, "get user modify info")
	}
	return &userInfo, nil
}

func (d *UserDao) GetInfoByUsername(ctx context.Context, username string) (*foundationview.UserInfo, error) {
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

func (d *UserDao) GetUserAccountInfo(ctx context.Context, id int) (*foundationview.UserAccountInfo, error) {
	var userAccountInfo foundationview.UserAccountInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname").
		Where("id = ?", id).
		Take(&userAccountInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "get user account infos")
	}
	return &userAccountInfo, nil
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

func (d *UserDao) GetUserAccountInfosByUsername(
	ctx context.Context,
	usernames []string,
) ([]*foundationview.UserAccountInfo, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var userAccountInfos []*foundationview.UserAccountInfo
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname").
		Where("username IN ?", usernames).
		Find(&userAccountInfos).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get user account infos")
	}
	return userAccountInfos, nil
}

func (d *UserDao) GetUserIdByUsername(ctx context.Context, username string) (int, error) {
	var userId int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("username = ?", username).
		Pluck("id", &userId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // User not found
		}
		return 0, metaerror.Wrap(err, "get user id by username")
	}
	return userId, nil
}

func (d *UserDao) GetUserIdsByUsername(ctx context.Context, usernames []string) ([]int, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var userIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("username IN ?", usernames).
		Pluck("id", &userIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get user ids by username")
	}
	return userIds, nil
}

func (d *UserDao) GetEmail(ctx context.Context, id int) (*string, error) {
	var email string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("id = ?", id).
		Pluck("email", &email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "get email by user id")
	}
	return &email, nil
}

func (d *UserDao) GetEmailByUsername(ctx context.Context, username string) (*string, error) {
	var email string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("username = ?", username).
		Pluck("email", &email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "get email by username")
	}
	return &email, nil
}

func (d *UserDao) GetRankAcAll(ctx context.Context, page int, size int) ([]*foundationview.UserRank, int, error) {
	var userRanks []*foundationview.UserRank
	var total int64
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("accept > 0").
		Count(&total).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "count total users")
	}
	err = d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Select("id, username, nickname, email, slogan, accept, attempt").
		Where("accept > 0").
		Order("accept DESC, attempt ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&userRanks).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "get user ranks")
	}
	return userRanks, int(total), nil
}

func (d *UserDao) FilterValidUserIds(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var validIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.User{}).
		Where("id IN ?", ids).
		Pluck("id", &validIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "filter valid user ids")
	}
	return validIds, nil
}

func (d *UserDao) UpdatePassword(ctx context.Context, username string, encodePassword string, nowTime time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("username = ?", username).
		Update("password", encodePassword).
		Update("modify_time", nowTime)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user password")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) UpdateUserInfo(
	ctx context.Context,
	userId int,
	request *foundationrequest.UserModifyInfo,
	modifyTime time.Time,
) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})

	gender := foundationenum.GetUserGender(request.Gender)

	res := db.Where("id = ?", userId).
		Updates(
			map[string]interface{}{
				"nickname":     request.Nickname,
				"slogan":       request.Slogan,
				"real_name":    request.RealName,
				"gender":       gender,
				"organization": request.Organization,
				"modify_time":  modifyTime,
			},
		)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user info")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
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

func (d *UserDao) UpdatePasswordByUserId(ctx *gin.Context, id int, encodePassword string, nowTime time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("id = ?", id).
		Update("password", encodePassword).
		Update("modify_time", nowTime)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user password")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) UpdateUserVjudgeUsername(ctx *gin.Context, id int, vjudgeId string, now time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("id = ?", id).
		Updates(
			map[string]interface{}{
				"vjudge_id":   vjudgeId,
				"modify_time": now,
			},
		)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user vjudge username")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}

func (d *UserDao) UpdateUserEmail(ctx *gin.Context, id int, email string, now time.Time) error {
	db := d.db.WithContext(ctx).Model(&foundationmodel.User{})
	res := db.Where("id = ?", id).
		Updates(
			map[string]interface{}{
				"email":       email,
				"modify_time": now,
			},
		)
	if res.Error != nil {
		return metaerror.Wrap(res.Error, "update user email")
	}
	if res.RowsAffected == 0 {
		return metaerror.New("no rows affected, user not found")
	}
	return nil
}
