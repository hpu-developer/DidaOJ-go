package service

import (
	"context"
	foundationdao "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	migratedao "migrate/dao"
	"strconv"
	"time"
)

type MigrateUserEojService struct {
}

// GORM 模型定义
type EojUser struct {
	Id         int       `gorm:"column:id"`
	Password   string    `gorm:"column:password"`
	DateJoined time.Time `gorm:"column:date_joined"`
	Username   string    `gorm:"column:username"`
	Email      string    `gorm:"column:email"`
	School     string    `gorm:"column:school"`
	Name       string    `gorm:"column:name"`
	StudentId  string    `gorm:"column:student_id"`
	Motto      string    `gorm:"column:motto"`
}

func (EojUser) TableName() string {
	return "account_user"
}

var singletonMigrateUserEojService = singleton.Singleton[MigrateUserEojService]{}

func GetMigrateUserEojService() *MigrateUserEojService {
	return singletonMigrateUserEojService.GetInstance(
		func() *MigrateUserEojService {
			return &MigrateUserEojService{}
		},
	)
}

func (s *MigrateUserEojService) Start() error {
	ctx := context.Background()

	// 初始化 GORM 客户端
	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var userModels []EojUser
	if err := eojDb.
		Model(&EojUser{}).
		Find(&userModels).Error; err != nil {
		return metaerror.Wrap(err, "query failed")
	}

	for _, userModel := range userModels {
		user := foundationmodel.NewUserBuilder().
			Username(userModel.Username).
			Nickname(userModel.Name).
			Number(&userModel.StudentId).
			Password(userModel.Password).
			Email(userModel.Email).
			CheckinCount(0).
			Slogan(&userModel.Motto).
			Organization(&userModel.School).
			RegTime(userModel.DateJoined).
			Accept(0).
			Attempt(0).
			Build()

		userMongo, err := foundationdao.GetUserDao().GetUserByUsername(ctx, user.Username)
		if err != nil {
			return err
		}
		var newId int
		if userMongo != nil {
			if user.Nickname != "" {
				userMongo.Nickname = user.Nickname
			}
			if user.Number != nil {
				userMongo.Number = user.Number
			}
			if user.Email != "" {
				userMongo.Email = user.Email
			}
			if user.Slogan != nil {
				userMongo.Slogan = user.Slogan
			}
			if user.Organization != nil {
				userMongo.Organization = user.Organization
			}
			err := foundationdao.GetUserDao().UpdateUser(ctx, userMongo.Id, userMongo)
			if err != nil {
				return err
			}
			newId = userMongo.Id
		} else {
			err = foundationdao.GetUserDao().InsertUser(ctx, user)
			if err != nil {
				return metaerror.Wrap(err, "insert user failed")
			}
			newId = user.Id
		}
		err = migratedao.GetMigrateMarkDao().Mark(
			ctx,
			"eoj-user",
			strconv.Itoa(userModel.Id),
			strconv.Itoa(newId),
		)
	}
	return nil
}
