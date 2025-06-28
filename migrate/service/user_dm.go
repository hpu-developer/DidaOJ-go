package service

import (
	"context"
	foundationdao "foundation/foundation-dao-mongo"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model-mongo"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	migratedao "migrate/dao"
	"strconv"
	"time"
)

type MigrateUserDmojService struct {
}

// GORM 模型定义
type DmojUser struct {
	UUID        string    `gorm:"column:uuid"`
	Username    string    `gorm:"column:username"`
	Password    string    `gorm:"column:password"`
	Nickname    string    `gorm:"column:nickname"`
	School      string    `gorm:"column:school"`
	Number      string    `gorm:"column:number"`
	Realname    string    `gorm:"column:realname"`
	Gender      string    `gorm:"column:gender"`
	Github      string    `gorm:"column:github"`
	CfUsername  string    `gorm:"column:cf_username"`
	Email       string    `gorm:"column:email"`
	Signature   string    `gorm:"column:signature"`
	GmtCreated  time.Time `gorm:"column:gmt_create"`
	GmtModified time.Time `gorm:"column:gmt_modified"`
}

func (DmojUser) TableName() string {
	return "user_info"
}

var singletonMigrateUserDmojService = singleton.Singleton[MigrateUserDmojService]{}

func GetMigrateUserDmojService() *MigrateUserDmojService {
	return singletonMigrateUserDmojService.GetInstance(
		func() *MigrateUserDmojService {
			return &MigrateUserDmojService{}
		},
	)
}

func (s *MigrateUserDmojService) Start(ctx context.Context) error {

	// 初始化 GORM 客户端
	dmojDb := metamysql.GetSubsystem().GetClient("dmoj")

	var userModels []DmojUser
	if err := dmojDb.
		Model(&DmojUser{}).
		Find(&userModels).Error; err != nil {
		return metaerror.Wrap(err, "query failed")
	}

	for _, userModel := range userModels {
		gender := foundationenum.UserGenderUnknown
		if userModel.Gender == "male" {
			gender = foundationenum.UserGenderMale
		} else if userModel.Gender == "female" {
			gender = foundationenum.UserGenderFemale
		}
		if userModel.Nickname == "" {
			userModel.Nickname = userModel.Username
		}

		user := foundationmodel.NewUserBuilder().
			Username(userModel.Username).
			Nickname(userModel.Nickname).
			Number(&userModel.Number).
			Password(userModel.Password).
			Email(userModel.Email).
			Slogan(&userModel.Signature).
			Organization(&userModel.School).
			RegTime(userModel.GmtCreated).
			RealName(&userModel.Realname).
			Gender(gender).
			Github(&userModel.Github).
			Codeforces(&userModel.CfUsername).
			Accept(0).
			Attempt(0).
			Build()

		user.ModifyTime = userModel.GmtModified

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
			if user.RealName != nil {
				userMongo.RealName = user.RealName
			}
			if user.Github != nil {
				userMongo.Github = user.Github
			}
			if user.Codeforces != nil {
				userMongo.Codeforces = user.Codeforces
			}
			if userMongo.RegTime == (time.Time{}) {
				userMongo.RegTime = user.RegTime
			}
			if userMongo.ModifyTime == (time.Time{}) {
				userMongo.ModifyTime = user.ModifyTime
			}
			userMongo.Gender = user.Gender
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
			"dmoj-user",
			userModel.UUID,
			strconv.Itoa(newId),
		)
	}
	return nil
}
