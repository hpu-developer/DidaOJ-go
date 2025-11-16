package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	foundationrun "foundation/foundation-run"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
)

type RunJobDao struct {
	db *gorm.DB
}

var singletonRunJobDao = singleton.Singleton[RunJobDao]{}

func GetRunJobDao() *RunJobDao {
	return singletonRunJobDao.GetInstance(
		func() *RunJobDao {
			dao := &RunJobDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

// AddRunJob 添加运行任务
func (d *RunJobDao) AddRunJob(ctx context.Context, runJob *foundationmodel.RunJob) error {
	return d.db.WithContext(ctx).Create(runJob).Error
}

// GetRunJob 获取运行任务
func (d *RunJobDao) GetRunJob(ctx context.Context, id int, userId int) (*foundationview.RunJob, error) {
	var runJob foundationview.RunJob
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.RunJob{}).
		Where("id = ?", id).
		Where("inserter = ?", userId).
		First(&runJob).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有找到记录
		}
		return nil, err
	}
	return &runJob, nil
}

// RequestRunJobListPending 获取待本地评测的 RunJob 列表，优先取最小的
func (d *RunJobDao) RequestRunJobListPending(
	ctx context.Context,
	maxCount int,
	judger string,
) ([]*foundationmodel.RunJob, error) {
	var jobs []*foundationmodel.RunJob

	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. SELECT ... FOR UPDATE SKIP LOCKED
			var jobIds []struct {
				Id int `gorm:"column:id"`
			}

			execSql := `
			SELECT j.id
			FROM run_job AS j
			WHERE j.status = ?
			ORDER BY j.status, j.id
			LIMIT ? FOR UPDATE SKIP LOCKED
		`
			if err := tx.Raw(execSql, foundationrun.RunStatusInit, maxCount).Scan(&jobIds).Error; err != nil {
				return err
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
			if err := tx.Model(&foundationmodel.RunJob{}).
				Where("id IN ?", ids).
				Updates(
					map[string]interface{}{
						"status": foundationrun.RunStatusQueuing,
						"judger": judger,
					},
				).Error; err != nil {
				return err
			}

			// 3. 返回完整任务信息
			return tx.Where("id IN ?", ids).Find(&jobs).Error
		},
	)

	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (d *RunJobDao) StartProcessRunJob(ctx context.Context, id int, judger string) (bool, error) {
	tx := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Where("id = ? AND judger = ?", id, judger).
		Update("status", foundationrun.RunStatusCompiling)
	if tx.Error != nil {
		return false, metaerror.Wrap(tx.Error, "failed to update job")
	}
	if tx.RowsAffected == 0 {
		// 没有匹配到符合条件的记录
		return false, nil
	}
	return true, nil
}

func (d *RunJobDao) MarkRunJobRunStatus(
	ctx context.Context,
	id int,
	judger string,
	status foundationrun.RunStatus,
	content string,
	time int,
	memory int,
) error {
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.RunJob{}).
		Where("id = ? AND judger = ?", id, judger).
		Updates(
			map[string]interface{}{
				"status":  status,
				"content": content,
				"time":    time,
				"memory":  memory,
			},
		).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to mark Run job status")
	}
	return nil
}
