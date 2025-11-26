package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
)

type BotGameDao struct {
	db *gorm.DB
}

var singletonBotDao = singleton.Singleton[BotGameDao]{}

func GetBotGameDao() *BotGameDao {
	return singletonBotDao.GetInstance(
		func() *BotGameDao {
			dao := &BotGameDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *BotGameDao) GetJudgeCode(ctx context.Context, gameId int) (string, error) {
	var judge string
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotGame{}).Where("id = ?", gameId).Pluck("judge_code", &judge).Error; err != nil {
		return "", metaerror.Wrap(err, "failed to get bot judge code")
	}
	return judge, nil
}
