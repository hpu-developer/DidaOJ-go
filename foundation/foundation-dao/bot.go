package foundationdao

import (
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
)

type BotDao struct {
	db *gorm.DB
}

var singletonBotDao = singleton.Singleton[BotDao]{}

func GetBotDao() *BotDao {
	return singletonBotDao.GetInstance(
		func() *BotDao {
			dao := &BotDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}
