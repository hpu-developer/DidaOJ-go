package foundationdao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"strings"
	"time"
)

type JudgeJobDao struct {
	db *gorm.DB
}

var singletonJudgeJobDao = singleton.Singleton[JudgeJobDao]{}

func GetJudgeJobDao() *JudgeJobDao {
	return singletonJudgeJobDao.GetInstance(
		func() *JudgeJobDao {
			dao := &JudgeJobDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *JudgeJobDao) GetJudgeJob(ctx context.Context, judgeId int, fields []string) (
	*foundationview.JudgeJob,
	error,
) {
	var view foundationview.JudgeJob
	var selectFields []string
	if len(fields) > 0 {
		selectFields = make([]string, 0, len(fields)+3)
		for _, field := range fields {
			selectFields = append(selectFields, "j."+field)
		}
		selectFields = append(
			selectFields,
			"u.username AS inserter_username",
			"u.nickname AS inserter_nickname",
			"judger.name AS judger_name",
			"jc.message AS compile_message",
		)
	} else {
		selectFields = []string{
			"j.*",
			"u.username AS inserter_username",
			"u.nickname AS inserter_nickname",
			"judger.name AS judger_name",
			"jc.message AS compile_message",
		}
	}
	err := d.db.WithContext(ctx).Table("judge_job AS j").
		Select(strings.Join(selectFields, ", ")).
		Joins("LEFT JOIN user AS u ON u.id = j.inserter").
		Joins("LEFT JOIN judger AS judger ON judger.key = j.judger").
		Joins("LEFT JOIN judge_job_compile AS jc ON jc.id = j.id").
		Where("j.id = ?", judgeId).
		Scan(&view).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to query judge job")
	}
	return &view, nil
}

func (d *JudgeJobDao) GetJudgeJobList(
	ctx context.Context,
	contestId int,
	problemId int,
	searchUserId int,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
	page int,
	pageSize int,
) ([]*foundationview.JudgeJob, error) {

	selectSql := `
			j.id, j.insert_time, j.language, j.score, j.status,
			j.time, j.memory, j.problem_id, j.inserter, j.code_length,
			u.username AS inserter_username, u.nickname AS inserter_nickname`

	if contestId > 0 {
		selectSql += ", cp.`index` AS contest_problem_index"
	}

	db := d.db.WithContext(ctx).Table("judge_job AS j").
		Select(
			selectSql,
		).
		Joins("LEFT JOIN user AS u ON u.id = j.inserter")
	if contestId > 0 {
		db = db.Joins(
			`
			LEFT JOIN contest_problem AS cp ON cp.id = j.contest_id AND cp.problem_id = j.problem_id
		`,
		)
		db = db.Where("j.contest_id = ?", contestId)
	} else {
		db = db.Where("j.contest_id IS NULL")
	}
	if problemId > 0 {
		db = db.Where("j.problem_id = ?", problemId)
	}
	if searchUserId > 0 {
		db = db.Where("j.inserter = ?", searchUserId)
	}
	if foundationjudge.IsValidJudgeLanguage(int(language)) {
		db = db.Where("j.language = ?", language)
	}
	if foundationjudge.IsValidJudgeStatus(int(status)) {
		db = db.Where("j.status = ?", status)
	}
	offset := (page - 1) * pageSize
	db = db.Order("j.id DESC").Limit(pageSize).Offset(offset)
	var list []*foundationview.JudgeJob
	if err := db.Scan(&list).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to query judge job list")
	}
	return list, nil
}

func (d *JudgeJobDao) GetJudgeTaskList(ctx *gin.Context, id int) ([]*foundationmodel.JudgeTask, error) {
	var tasks []*foundationmodel.JudgeTask
	err := d.db.WithContext(ctx).Model(&foundationmodel.JudgeTask{}).
		Where("id = ?", id).
		Order("task_id ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to query judge task list")
	}
	return tasks, nil
}

func (d *JudgeJobDao) GetProblemAttemptStatus(
	ctx context.Context, authorId int, problemIds []int,
	contestId int, startTime *time.Time, endTime *time.Time,
) (map[int]foundationenum.ProblemAttemptStatus, error) {
	if len(problemIds) == 0 {
		return nil, nil
	}
	type Result struct {
		ProblemId  int
		HasAC      int
		HasAttempt int
	}
	db := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{}).
		Select(
			"problem_id, MAX(CASE WHEN status = ? THEN 1 ELSE 0 END) AS has_ac, MAX(CASE WHEN status != ? THEN 1 ELSE 0 END) AS has_attempt",
			foundationjudge.JudgeStatusAC, foundationjudge.JudgeStatusAC,
		).
		Where("inserter = ?", authorId).
		Where("problem_id IN ?", problemIds)
	if contestId > 0 {
		db = db.Where("contest_id = ?", contestId)
	}
	if startTime != nil {
		db = db.Where("insert_time >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("insert_time <= ?", *endTime)
	}
	db = db.Group("problem_id")
	var results []Result
	if err := db.Scan(&results).Error; err != nil {
		return nil, metaerror.Wrap(err, "failed to query judge job")
	}
	statusMap := make(map[int]foundationenum.ProblemAttemptStatus, len(problemIds))
	for _, r := range results {
		switch {
		case r.HasAC > 0:
			statusMap[r.ProblemId] = foundationenum.ProblemAttemptStatusAccepted
		case r.HasAttempt > 0:
			statusMap[r.ProblemId] = foundationenum.ProblemAttemptStatusAttempt
		}
	}
	return statusMap, nil
}

func (d *JudgeJobDao) GetUserAcProblemIds(db *gorm.DB, userId int) ([]string, error) {
	var problemIds []string
	err := db.Model(&foundationmodel.JudgeJob{}).
		Select("DISTINCT problem_id").
		Where("status = ?", foundationjudge.JudgeStatusAC).
		Where("inserter = ?", userId).
		Pluck("problem_id", &problemIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to get distinct problem_ids")
	}
	return problemIds, nil
}

func (d *JudgeJobDao) GetAcUserIds(db *gorm.DB, problemId int, limit int) ([]int, error) {
	var acUserIds []int
	subDb := db.Model(&foundationmodel.JudgeJob{}).
		Select("DISTINCT inserter").
		Where("status = ?", foundationjudge.JudgeStatusAC)
	if problemId > 0 {
		subDb = subDb.Where("problem_id = ?", problemId)
	}
	subDb = subDb.Limit(1000)
	if err := subDb.Pluck("inserter", &acUserIds).Error; err != nil {
		return nil, err
	}
	return acUserIds, nil
}

func (d *JudgeJobDao) GetProblemRecommendByProblem(
	ctx context.Context,
	userId int,
	hasAuth bool,
	problemId int,
) ([]int, error) {
	db := d.db.WithContext(ctx)
	userAcProblems, err := d.GetUserAcProblemIds(db, userId)
	if err != nil {
		return nil, err
	}
	acUserIDs, err := d.GetAcUserIds(db, problemId, 1000)
	if err != nil {
		return nil, err
	}
	if len(acUserIDs) == 0 {
		return nil, nil
	}
	type Result struct {
		ProblemId int
		Count     int
	}
	var recResults []Result

	recQuery := db.Table("judge_job AS jj").
		Select("jj.problem_id, COUNT(*) AS count").
		Joins("JOIN problem p ON p.id = jj.problem_id").
		Where("jj.status = ?", foundationjudge.JudgeStatusAC).
		Where("jj.insert_time IS NOT NULL").
		Where("jj.inserter IN ?", acUserIDs).
		Where("jj.problem_id NOT IN ?", userAcProblems)

	if problemId > 0 {
		recQuery = recQuery.Where("jj.problem_id != ?", problemId)
	}

	if !hasAuth {
		if userId > 0 {
			recQuery = recQuery.Where(
				`
				(p.private = 0
				OR p.inserter = ?
				OR p.id IN (SELECT problem_id FROM problem_member WHERE user_id = ?)
				OR p.id IN (SELECT problem_id FROM problem_member_auth WHERE user_id = ?))`,
				userId, userId, userId,
			)
		} else {
			recQuery = recQuery.Where("p.private = 0")
		}
	}

	recQuery = recQuery.Group("jj.problem_id").
		Order("count DESC").
		Limit(20)

	if err := recQuery.Scan(&recResults).Error; err != nil {
		return nil, err
	}
	if len(recResults) == 0 {
		return nil, nil
	}

	finalIds := make([]int, 0, len(recResults))
	for _, r := range recResults {
		finalIds = append(finalIds, r.ProblemId)
	}
	return finalIds, nil
}

func (d *JudgeJobDao) GetRankAcProblem(
	ctx context.Context,
	approveStartTime *time.Time,
	approveEndTime *time.Time,
	page int,
	pageSize int,
) ([]*foundationview.UserRank, int, error) {
	db := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{})
	db = db.Where("status = ?", foundationjudge.JudgeStatusAC)
	if approveStartTime != nil {
		db = db.Where("insert_time >= ?", *approveStartTime)
	}
	if approveEndTime != nil {
		db = db.Where("insert_time < ?", *approveEndTime)
	}
	subQuery := db.
		Select("inserter AS id, COUNT(DISTINCT problem_id) AS problem_count").
		Group("inserter")
	var result []*foundationview.UserRank
	err := d.db.Table("(?) AS t", subQuery).
		Select("t.id, t.problem_count, u.username, u.nickname, u.slogan").
		Joins("LEFT JOIN user u ON u.id = t.id").
		Order("t.problem_count DESC, t.id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&result).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to query user rank")
	}
	var total int64
	err = d.db.Table("(?) AS t", subQuery).Count(&total).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count user rank total")
	}
	return result, int(total), nil
}

func (d *JudgeJobDao) GetContestRanks(
	ctx context.Context,
	id int,
	lockTime *time.Time,
	problemMap map[int]uint8,
) ([]*foundationview.ContestRank, error) {

	var execSql string
	var rows *sql.Rows
	var err error

	if lockTime == nil {
		execSql = `
SELECT inserter,
       u.username AS username,
       u.nickname AS nickname,
       JSON_ARRAYAGG(
               JSON_OBJECT(
                       'id', problem_id,
                       'attempt', count,
                       'ac', DATE_FORMAT(ac, '%Y-%m-%dT%H:%i:%sZ')
               )
       )          AS problems
FROM (SELECT j.inserter,
             j.problem_id,
             COUNT(*)       AS count,
             ac.insert_time AS ac
      FROM judge_job j
               LEFT JOIN (SELECT inserter, problem_id, MIN(id) AS ac_id
                          FROM judge_job
                          WHERE contest_id = ?
                            AND status = ?
                          GROUP BY inserter, problem_id) fa ON j.inserter = fa.inserter AND j.problem_id = fa.problem_id
               LEFT JOIN judge_job ac ON ac.id = fa.ac_id
      WHERE j.contest_id = ?
        AND (fa.ac_id IS NULL OR j.id < fa.ac_id)
      GROUP BY j.inserter, j.problem_id) AS flat
         LEFT JOIN user as u ON flat.inserter = u.id
GROUP BY inserter
`
		rows, err = d.db.WithContext(ctx).Raw(execSql, id, foundationjudge.JudgeStatusAC, id).Rows()
	} else {
		execSql = `
SELECT inserter,
       u.username AS username,
       u.nickname AS nickname,
       JSON_ARRAYAGG(
               JSON_OBJECT(
                       'id', problem_id,
                       'attempt', count_before,
                       'lock', count_after,
                       'ac', DATE_FORMAT(ac, '%Y-%m-%dT%H:%i:%sZ')
               )
       )          AS problems
FROM (SELECT j.inserter,
             j.problem_id,
             SUM(j.insert_time < ?)  AS count_before,
             SUM(j.insert_time >= ?) AS count_after,
             ac.insert_time                   AS ac
      FROM judge_job j
               LEFT JOIN (SELECT inserter, problem_id, MIN(id) AS ac_id
                          FROM judge_job
                          WHERE contest_id = ?
                            AND status = ?
                            AND insert_time < ?
                          GROUP BY inserter, problem_id) fa ON j.inserter = fa.inserter AND j.problem_id = fa.problem_id
               LEFT JOIN judge_job ac ON ac.id = fa.ac_id
      WHERE j.contest_id = ?
        AND (fa.ac_id IS NULL OR j.id < fa.ac_id)
      GROUP BY j.inserter, j.problem_id) AS flat
         LEFT JOIN user u ON flat.inserter = u.id
GROUP BY inserter;`

		rows, err = d.db.WithContext(ctx).Debug().Raw(
			execSql,
			lockTime,
			lockTime,
			id,
			foundationjudge.JudgeStatusAC,
			lockTime,
			id,
		).Rows()
	}

	if err != nil {
		return nil, metaerror.Wrap(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close rows,id:%d"), id)
		}
	}(rows)

	var ranks []*foundationview.ContestRank

	for rows.Next() {
		var rank foundationview.ContestRank
		var jsonProblems json.RawMessage
		err := rows.Scan(&rank.Inserter, &rank.InserterUsername, &rank.InserterNickname, &jsonProblems)
		if err != nil {
			return nil, metaerror.Wrap(err, "failed to scan row")
		}
		err = json.Unmarshal(jsonProblems, &rank.Problems)
		if err != nil {
			return nil, metaerror.Wrap(err, "failed to unmarshal problems")
		}
		for _, problem := range rank.Problems {
			problem.Index = problemMap[problem.Id]
			problem.Id = 0
		}
		ranks = append(ranks, &rank)
	}
	return ranks, nil
}

// RequestJudgeJobListPendingJudge 获取待评测的 JudgeJob 列表，优先取最小的
func (d *JudgeJobDao) RequestJudgeJobListPendingJudge(
	ctx context.Context,
	maxCount int,
	judger string,
) ([]*foundationmodel.JudgeJob, error) {
	var jobs []*foundationmodel.JudgeJob

	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where(
					"status IN ?", []foundationjudge.JudgeStatus{
						foundationjudge.JudgeStatusInit,
						foundationjudge.JudgeStatusRejudge,
					},
				).
				Order("status ASC, id ASC").
				Limit(maxCount).
				Find(&jobs).Error
			if err != nil {
				return err
			}
			if len(jobs) == 0 {
				return gorm.ErrRecordNotFound
			}

			// 批量更新每条记录的状态
			for _, job := range jobs {
				err := tx.Model(job).Updates(
					map[string]interface{}{
						"status":     foundationjudge.JudgeStatusQueuing,
						"judger":     judger,
						"judge_time": time.Now(),
					},
				).Error
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有待评测的任务
		}
		return nil, metaerror.Wrap(err, "failed to request judge job list pending judge")
	}
	return jobs, nil
}

func (d *JudgeJobDao) StartProcessJudgeJob(ctx context.Context, id int, judger string) (bool, error) {
	tx := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Where("id = ? AND judger = ?", id, judger).
		Update("status", foundationjudge.JudgeStatusCompiling)
	if tx.Error != nil {
		return false, metaerror.Wrap(tx.Error, "failed to update job")
	}
	if tx.RowsAffected == 0 {
		// 没有匹配到符合条件的记录
		return false, nil
	}
	return true, nil
}

func (d *JudgeJobDao) MarkJudgeJobJudgeStatus(
	ctx context.Context,
	id int,
	judger string,
	status foundationjudge.JudgeStatus,
) error {
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Where("id = ? AND judger = ?", id, judger).
		Update("status", status).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to mark judge job status")
	}
	return nil
}

func (d *JudgeJobDao) MarkJudgeJobTaskTotal(ctx context.Context, id int, judger string, taskTotalCount int) error {
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Where("id = ? AND judger = ?", id, judger).
		Update("task_current", 0).
		Update("task_total", taskTotalCount).Error
	if err != nil {
		return metaerror.Wrap(err, "failed to mark judge job task total")
	}
	return nil
}

func (d *JudgeJobDao) AddJudgeJobTaskCurrent(
	ctx context.Context,
	id int,
	judger string,
	task *foundationmodel.JudgeTask,
) error {
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 确保 judge_job 中有这条记录且 judger 匹配
			var job foundationmodel.JudgeJob
			if err := tx.
				Where("id = ? AND judger = ?", id, judger).
				First(&job).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("judge_job not found with id=%d and judger=%s", id, judger)
				}
				return metaerror.Wrap(err, "failed to find judge_job")
			}
			// 插入任务记录（保底设置 id 关联）
			task.Id = id
			if err := tx.Create(task).Error; err != nil {
				return metaerror.Wrap(err, "failed to insert judge_task")
			}
			// 更新 task_current 计数器
			if err := tx.Model(&foundationmodel.JudgeJob{}).
				Where("id = ?", id).
				UpdateColumn("task_current", gorm.Expr("task_current + 1")).Error; err != nil {
				return metaerror.Wrap(err, "failed to increment task_current")
			}
			return nil
		},
	)
}

func (d *JudgeJobDao) MarkJudgeJobJudgeFinalStatus(
	ctx context.Context, id int, judger string,
	status foundationjudge.JudgeStatus,
	problemId int,
	userId int,
	score int,
	time int,
	memory int,
) error {
	markStatusFunc := func(tx *gorm.DB) error {
		// 限定条件 id + judger，避免误更新其他评测
		res := tx.Model(&foundationmodel.JudgeJob{}).
			Where("id = ? AND judger = ?", id, judger).
			Updates(
				map[string]interface{}{
					"status": status,
					"score":  score,
					"time":   time,
					"memory": memory,
				},
			)

		if res.Error != nil {
			return metaerror.Wrap(res.Error, "failed to mark judge job status")
		}
		if res.RowsAffected == 0 {
			return metaerror.New("no judge_job found with id=%d and judger=%s", id, judger)
		}
		return nil
	}

	if status == foundationjudge.JudgeStatusAC {
		// 事务中进行多个表的更新
		return d.db.WithContext(ctx).Transaction(
			func(tx *gorm.DB) error {
				if err := markStatusFunc(tx); err != nil {
					return err
				}

				// problem 表 accept++
				if err := tx.Model(&foundationmodel.Problem{}).
					Where("id = ?", problemId).
					UpdateColumn("accept", gorm.Expr("accept + ?", 1)).Error; err != nil {
					return metaerror.Wrap(err, "failed to update problem accept count")
				}

				// user 表 accept++
				if err := tx.Model(&foundationmodel.User{}).
					Where("id = ?", userId).
					UpdateColumn("accept", gorm.Expr("accept + ?", 1)).Error; err != nil {
					return metaerror.Wrap(err, "failed to update user accept count")
				}

				return nil
			},
		)
	} else {
		// 非 AC 情况下，只更新 judge_job 状态
		return markStatusFunc(d.db.WithContext(ctx))
	}
}
func (d *JudgeJobDao) RejudgeJob(ctx context.Context, id int) error {
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. 加锁查找 judge_job（防止并发修改）
			var job struct {
				ID        int                         `gorm:"column:id"`
				ProblemId int                         `gorm:"column:problem_id"`
				Inserter  int                         `gorm:"column:inserter"`
				Status    foundationjudge.JudgeStatus `gorm:"column:status"`
			}
			if err := tx.Table("judge_job").
				Select("id, problem_id, inserter, status").
				Where("id = ?", id).
				Clauses(clause.Locking{Strength: "UPDATE"}). // 加锁
				First(&job).Error; err != nil {
				return metaerror.Wrap(err, "find judge_job error")
			}

			// 2. 计算更新偏移
			problemAcceptDelta := 0
			userAcceptDelta := 0
			if job.Status == foundationjudge.JudgeStatusAC {
				problemAcceptDelta--
				userAcceptDelta--
			}

			// 3. 更新 judge_job
			updateMap := map[string]interface{}{
				"status": foundationjudge.JudgeStatusRejudge,
				"score":  nil, "time": nil, "memory": nil,
				"task_current": nil,
				"task_total":   nil,
				"judger":       nil,
				"judge_time":   nil,
			}
			if err := tx.Table("judge_job").
				Where("id = ?", id).
				Updates(updateMap).Error; err != nil {
				return metaerror.Wrap(err, "failed to update judge_job")
			}

			// 4. 删除 judge_job_compile 中对应记录
			if err := tx.Table("judge_job_compile").
				Where("id = ?", id).
				Delete(nil).Error; err != nil {
				return metaerror.Wrap(err, "failed to delete compile message")
			}

			// 5. 删除 judge_task 中对应记录
			if err := tx.Table("judge_task").
				Where("id = ?", id).
				Delete(nil).Error; err != nil {
				return metaerror.Wrap(err, "failed to delete judge_task")
			}

			// 6. 更新 problem.accept
			if problemAcceptDelta != 0 {
				if err := tx.Table("problem").
					Where("id = ?", job.ProblemId).
					Update("accept", gorm.Expr("accept + ?", problemAcceptDelta)).Error; err != nil {
					return metaerror.Wrap(err, "failed to update problem accept count")
				}
			}

			// 7. 更新 user.accept
			if userAcceptDelta != 0 {
				if err := tx.Table("user").
					Where("id = ?", job.Inserter).
					Update("accept", gorm.Expr("accept + ?", userAcceptDelta)).Error; err != nil {
					return metaerror.Wrap(err, "failed to update user accept count")
				}
			}
			return nil
		},
	)
}

func (d *JudgeJobDao) InsertJudgeJob(
	ctx context.Context,
	judgeJob *foundationmodel.JudgeJob,
) error {
	if judgeJob == nil {
		return metaerror.New("judgeJob is nil")
	}
	if err := d.db.WithContext(ctx).Create(judgeJob).Error; err != nil {
		return metaerror.Wrap(err, "insert judgeJob")
	}
	return nil
}
