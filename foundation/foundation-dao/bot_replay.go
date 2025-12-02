package foundationdao

import (
	"context"
	foundationbot "foundation/foundation-bot"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	metatime "meta/meta-time"
	"meta/singleton"

	"gorm.io/gorm"
)

type BotReplayDao struct {
	db *gorm.DB
}

var singletonBotReplayDao = singleton.Singleton[BotReplayDao]{}

func GetBotReplayDao() *BotReplayDao {
	return singletonBotReplayDao.GetInstance(
		func() *BotReplayDao {
			dao := &BotReplayDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

// RequestBotReplayListPending 获取待处理的BotReplay任务列表
func (d *BotReplayDao) RequestBotReplayListPending(
	ctx context.Context,
	maxCount int,
	judger string,
) ([]*foundationmodel.BotReplay, error) {
	var jobs []*foundationmodel.BotReplay

	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. SELECT ... FOR UPDATE SKIP LOCKED
			var jobIds []struct {
				Id int `gorm:"column:id"`
			}

			execSql := `
			SELECT j.id
			FROM bot_replay AS j
			WHERE j.status = ?
			ORDER BY j.status, j.id
			LIMIT ? FOR UPDATE SKIP LOCKED
		`
			if err := tx.Raw(execSql, foundationbot.BotGameStatusInit, maxCount).Scan(&jobIds).Error; err != nil {
				return metaerror.Wrap(err, "failed to request bot replay list pending")
			}

			if len(jobIds) == 0 {
				return nil // 没有任务可领取
			}

			// 提取出 id 列表
			ids := make([]int, len(jobIds))
			for i, job := range jobIds {
				ids[i] = job.Id
			}

			// 2. UPDATE 任务状态
			if err := tx.Model(&foundationmodel.BotReplay{}).
				Where("id IN ?", ids).
				Updates(
					map[string]interface{}{
						"status":     foundationbot.BotGameStatusQueuing,
						"judger":     judger,
						"judge_time": metatime.GetTimeNow(),
					},
				).Error; err != nil {
				return err
			}

			// 3. 返回完整任务信息
			err := tx.Where("id IN ?", ids).Find(&jobs).Error
			if err != nil {
				return metaerror.Wrap(err, "failed to request bot replay list pending")
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (d *BotReplayDao) MarkBotReplayInfo(ctx context.Context, id int, judger string, info string) error {
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.BotReplay{}).
		Where("id = ? AND judger = ?", id, judger).
		Updates(
			map[string]interface{}{
				"status": foundationbot.BotGameStatusRunning,
				"info":   info,
			},
		).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to mark bot replay status")
	}
	return nil
}

func (d *BotReplayDao) MarkBotReplayParam(ctx context.Context, id int, judger string, param string) error {
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.BotReplay{}).
		Where("id = ? AND judger = ?", id, judger).
		Updates(
			map[string]interface{}{
				"param": param,
			},
		).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to mark bot replay param")
	}
	return nil
}

// MarkBotReplayRunStatus 更新BotReplay任务状态
func (d *BotReplayDao) MarkBotReplayRunStatus(
	ctx context.Context,
	id int,
	judger string,
	status foundationbot.BotGameStatus,
	message string,
) error {
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.BotReplay{}).
		Where("id = ? AND judger = ?", id, judger).
		Updates(
			map[string]interface{}{
				"status":  status,
				"message": message,
			},
		).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to mark bot replay status")
	}
	return nil
}
