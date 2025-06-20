package service

import (
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type MigrateUserDmService struct {
}

// GORM 模型定义
type DmUser struct {
	Id       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
	Nickname string `gorm:"column:nickname"`
	School   string `gorm:"column:school"`
	Course   string `gorm:"column:course"`
	Number   string `gorm:"column:number"`
	Realname string `gorm:"column:realname"`
	Gender   string `gorm:"column:gender"`
}

func (DmUser) TableName() string {
	return "account_user"
}

var singletonMigrateUserDmService = singleton.Singleton[MigrateUserDmService]{}

func GetMigrateUserDmService() *MigrateUserDmService {
	return singletonMigrateUserDmService.GetInstance(
		func() *MigrateUserDmService {
			return &MigrateUserDmService{}
		},
	)
}

func (s *MigrateUserDmService) Start() error {
	//ctx := context.Background()

	// 初始化 GORM 客户端
	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var userModels []EojUser
	if err := eojDb.
		Model(&EojUser{}).
		Find(&userModels).Error; err != nil {
		return metaerror.Wrap(err, "query failed")
	}

	//for _, userModel := range userModels {
	//
	//}
	return nil
}
