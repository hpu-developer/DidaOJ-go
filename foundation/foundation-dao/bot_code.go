package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
)

type BotCodeDao struct {
	db *gorm.DB
}

var singletonBotCodeDao = singleton.Singleton[BotCodeDao]{}

func GetBotCodeDao() *BotCodeDao {
	return singletonBotCodeDao.GetInstance(
		func() *BotCodeDao {
			dao := &BotCodeDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *BotCodeDao) GetBotCodes(ctx context.Context, botIds []int) ([]*foundationview.BotCodeView, error) {
	var botCodes []*foundationview.BotCodeView
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotCode{}).
		Where("id IN ?", botIds).
		Select("id, language, code, version, inserter").
		Find(&botCodes).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to get bot code map")
	}
	return botCodes, nil
}
