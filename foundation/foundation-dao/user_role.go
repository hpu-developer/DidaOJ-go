package foundationdao

import (
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
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *UserRoleDao) GetUserRoles(userId int) ([]string, error) {
	var roles []string
	err := d.db.Model(&UserRole{}).
		Where("user_id = ?", userId).
		Pluck("role", &roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
