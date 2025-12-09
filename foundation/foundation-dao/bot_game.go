package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
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

func (d *BotGameDao) CheckGameEditAuth(ctx context.Context, gameId int, userId int) (bool, error) {
	var exists int
	err := d.db.WithContext(ctx).Raw(
		`SELECT 1
		FROM bot_game p
		WHERE p.id = ? AND p.inserter = ?
		LIMIT 1
	`, gameId, userId,
	).Scan(&exists).Error
	if err != nil {
		return false, metaerror.Wrap(err, "check edit permission failed")
	}
	return exists == 1, nil
}

// GetBotGameIdByKey 根据key获取BotGameId
func (d *BotGameDao) GetBotGameIdByKey(ctx context.Context, gameKey string) (int, error) {
	var gameId int
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotGame{}).
		Select("id").
		Where("LOWER(game_key) = LOWER(?)", gameKey).
		Take(&gameId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, metaerror.Wrap(err, "failed to get bot game id by key")
	}
	return gameId, nil
}

// GetBotGameByKey 根据key获取BotGame
func (d *BotGameDao) GetBotGameByKey(ctx context.Context, gameKey string) (*foundationmodel.BotGame, error) {
	var botGame foundationmodel.BotGame
	if err := d.db.WithContext(ctx).Where("LOWER(game_key) = LOWER(?)", gameKey).First(&botGame).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "failed to get bot game by key")
	}
	return &botGame, nil
}

// GetBotGameDescription 获取游戏描述
func (d *BotGameDao) GetBotGameDescription(ctx context.Context, gameId int) (*string, error) {
	var result struct {
		Description string `gorm:"column:description"`
	}
	err := d.db.WithContext(ctx).Model(&foundationmodel.BotGame{}).
		Select("description").
		Where("id = ?", gameId).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &result.Description, err
}

// GetBotGameList 获取游戏列表
func (d *BotGameDao) GetBotGameList(ctx context.Context) ([]*foundationview.BotGameListView, error) {
	var botGameList []*foundationview.BotGameListView
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotGame{}).
		Select("id, game_key, title, introduction, player_min, player_max").
		Find(&botGameList).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to get bot game list")
	}
	return botGameList, nil
}

func (d *BotGameDao) GetJudgeCode(ctx context.Context, gameId int) (string, error) {
	var judge string
	if err := d.db.WithContext(ctx).Model(&foundationmodel.BotGame{}).Where("id = ?", gameId).Pluck("judge_code", &judge).Error; err != nil {
		return "", metaerror.Wrap(err, "failed to get bot judge code")
	}
	return judge, nil
}

func (d *BotGameDao) UpdateBotGame(ctx context.Context, gameId int, botGame *foundationmodel.BotGame) error {
	updateData := map[string]interface{}{
		"title":       botGame.Title,
		"description": botGame.Description,
		"judge_code":  botGame.JudgeCode,
		"modifier":    botGame.Modifier,
		"modify_time": botGame.ModifyTime,
	}
	txRes := d.db.WithContext(ctx).Model(&foundationmodel.BotGame{}).
		Where("id = ?", gameId).
		Updates(updateData)
	if txRes.Error != nil {
		return txRes.Error
	}
	return nil
}
