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

type BotAgentDao struct {
	db *gorm.DB
}

var singletonBotAgentDao = singleton.Singleton[BotAgentDao]{}

func GetBotAgentDao() *BotAgentDao {
	return singletonBotAgentDao.GetInstance(
		func() *BotAgentDao {
			dao := &BotAgentDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *BotAgentDao) GetBotAgentList(ctx context.Context, agentId int, name string, inserter int) ([]*foundationview.BotAgentView, error) {
	var botAgents []*foundationview.BotAgentView

	db := d.db.WithContext(ctx).Table("bot_agent as ba")
	if agentId > 0 {
		db = db.Where("ba.id = ?", agentId)
	}
	if name != "" {
		db = db.Where("ba.name LIKE ?", "%"+name+"%")
	}
	if inserter > 0 {
		db = db.Where("ba.inserter = ?", inserter)
	}
	if err := db.Select(`ba.id, ba.version, ba.inserter, ba.name, 
	u.username as inserter_username, u.nickname as inserter_nickname, u.email as inserter_email`).
		Joins("LEFT JOIN \"user\" as u ON u.id = ba.inserter").
		Find(&botAgents).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to get bot agent list")
	}
	return botAgents, nil
}

func (d *BotAgentDao) GetBotAgents(ctx context.Context, botIds []int) ([]*foundationview.BotAgentView, error) {
	var botCodes []*foundationview.BotAgentView
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotAgent{}).
		Where("id IN ?", botIds).
		Select("id, language, code, version, inserter, name").
		Find(&botCodes).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to get bot code map")
	}
	return botCodes, nil
}

func (d *BotAgentDao) GetBotPlayers(ctx context.Context, botIds []int) ([]*foundationview.BotAgentPlayerView, error) {
	var botPlayers []*foundationview.BotAgentPlayerView
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotAgent{}).
		Where("id IN ?", botIds).
		Select("id, name").
		Find(&botPlayers).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to get bot players")
	}
	return botPlayers, nil
}
