package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type UserRoleDao struct {
	db *gorm.DB
}

var singletonUserRoleDao = singleton.Singleton[UserRoleDao]{}

func GetUserRoleDao() *UserRoleDao {
	return singletonUserRoleDao.GetInstance(
		func() *UserRoleDao {
			dao := &UserRoleDao{}
			db := metamysql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model(&foundationmodel.UserRole{})
			return dao
		},
	)
}

func (d *UserRoleDao) GetUserRoles(ctx context.Context, userId int) ([]string, error) {
	var roles []string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.UserRole{}).
		Where("id = ?", userId).
		Pluck("role_id", &roles).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return roles, nil
}
