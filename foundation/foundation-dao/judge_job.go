package foundationdao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

func (d *JudgeJobDao) GetJudgeJobViewAuth(ctx context.Context, id int) (*foundationview.JudgeJobViewAuth, error) {
	var auth foundationview.JudgeJobViewAuth
	err := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{}).
		Select("id, contest_id, problem_id, inserter, private, inserter").
		Where("id = ?", id).
		First(&auth).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有找到记录
		}
		return nil, metaerror.Wrap(err, "failed to query judge job view auth")
	}
	return &auth, nil
}

func (d *JudgeJobDao) GetJudgeCode(ctx context.Context, id int) (foundationjudge.JudgeLanguage, *string, error) {
	var m struct {
		Language foundationjudge.JudgeLanguage
		Code     string
	}
	err := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{}).
		Select("language, code").
		Where("id = ?", id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return foundationjudge.JudgeLanguageUnknown, nil, nil // 没有找到记录
		}
		return foundationjudge.JudgeLanguageUnknown, nil, metaerror.Wrap(err, "failed to query judge code")
	}
	return m.Language, &m.Code, nil

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
			"p.`key` AS problem_key",
		)
	} else {
		selectFields = []string{
			"j.*",
			"u.username AS inserter_username",
			"u.nickname AS inserter_nickname",
			"judger.name AS judger_name",
			"jc.message AS compile_message",
			"p.`key` AS problem_key",
		}
	}

	err := d.db.WithContext(ctx).Table("judge_job AS j").
		Select(strings.Join(selectFields, ", ")).
		Joins("LEFT JOIN user AS u ON u.id = j.inserter").
		Joins("LEFT JOIN judger AS judger ON judger.key = j.judger").
		Joins("LEFT JOIN judge_job_compile AS jc ON jc.id = j.id").
		Joins("LEFT JOIN problem AS p ON p.id = j.problem_id").
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

	selectSql += ", p.`key` as problem_key"

	if contestId > 0 {
		selectSql += ", cp.`index` AS contest_problem_index"
	}

	db := d.db.WithContext(ctx).Table("judge_job AS j").
		Select(
			selectSql,
		).
		Joins("LEFT JOIN user AS u ON u.id = j.inserter").
		Joins("LEFT JOIN problem AS p ON p.id = j.problem_id")
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
	ctx context.Context, inserter int, problemIds []int,
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
		Where("inserter = ?", inserter).
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

func (d *JudgeJobDao) GetUserAcProblemIds(ctx context.Context, userId int) ([]*foundationview.ProblemViewKey, error) {
	var problemIds []*foundationview.ProblemViewKey
	err := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{}).
		Select("DISTINCT problem_id as id, p.`key` as `key`").
		Where("status = ?", foundationjudge.JudgeStatusAC).
		Where("inserter = ?", userId).
		Joins("JOIN problem AS p ON p.id = judge_job.problem_id").
		Pluck("problem_id", &problemIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to get distinct problem_ids")
	}
	return problemIds, nil
}

func (d *JudgeJobDao) GetAcUserIds(ctx context.Context, problemId int, limit int) ([]int, error) {
	var acUserIds []int
	subDb := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{}).
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
	userAcProblems, err := d.GetUserAcProblemIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	acUserIDs, err := d.GetAcUserIds(ctx, problemId, 1000)
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
	recQuery := d.db.WithContext(ctx).Debug().Table("judge_job AS jj").
		Select("jj.problem_id, COUNT(*) AS count").
		Joins("JOIN problem p ON p.id = jj.problem_id").
		Where("jj.status = ?", foundationjudge.JudgeStatusAC).
		Where("jj.inserter IN ?", acUserIDs)

	if len(userAcProblems) > 0 {
		recQuery = recQuery.Where("jj.problem_id NOT IN ?", userAcProblems)
	}
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
	insertStartTime *time.Time,
	insertEndTime *time.Time,
	page int,
	pageSize int,
) ([]*foundationview.UserRank, int, error) {
	db := d.db.WithContext(ctx).Model(&foundationmodel.JudgeJob{})
	db = db.Where("status = ?", foundationjudge.JudgeStatusAC)
	if insertStartTime != nil {
		db = db.Where("insert_time >= ?", *insertStartTime)
	}
	if insertEndTime != nil {
		db = db.Where("insert_time < ?", *insertEndTime)
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

func (d *JudgeJobDao) GetJudgeJobCountStaticsRecently(ctx context.Context) (
	[]*foundationview.JudgeJobCountStatics,
	error,
) {
	const days = 30
	end := time.Now()
	start := end.AddDate(0, 0, -days+1)

	type aggResult struct {
		Date   time.Time
		Status foundationjudge.JudgeStatus
		Count  int
	}

	var results []aggResult

	err := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Select(`DATE(insert_time) AS date, status, COUNT(*) AS count`).
		Where(
			"insert_time >= ? AND insert_time < ?",
			time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()),
			time.Date(end.Year(), end.Month(), end.Day()+1, 0, 0, 0, 0, end.Location()),
		).
		Group("date, status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 构造统计映射
	resultMap := make(map[string]*foundationview.JudgeJobCountStatics)
	for _, res := range results {
		dateStr := res.Date.Format("2006-01-02")
		if _, ok := resultMap[dateStr]; !ok {
			resultMap[dateStr] = &foundationview.JudgeJobCountStatics{
				Date:    res.Date,
				Accept:  0,
				Attempt: 0,
			}
		}
		stat := resultMap[dateStr]
		stat.Attempt += res.Count
		if res.Status == foundationjudge.JudgeStatusAC {
			stat.Accept += res.Count
		}
	}

	// 构造返回值，补齐没有数据的日期
	var statList []*foundationview.JudgeJobCountStatics
	for i := 0; i < days; i++ {
		date := start.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		if stat, ok := resultMap[dateStr]; ok {
			statList = append(statList, stat)
		} else {
			statList = append(
				statList, &foundationview.JudgeJobCountStatics{
					Date:    date,
					Accept:  0,
					Attempt: 0,
				},
			)
		}
	}

	return statList, nil
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
                       'ac', ac
               )
       )          AS problems
FROM (SELECT j.inserter,
             j.problem_id,
             SUM(fa.ac_id is null OR j.id < fa.ac_id) AS count,
             ac.insert_time                           AS ac
      FROM judge_job j
               LEFT JOIN (SELECT inserter, problem_id, MIN(id) AS ac_id
                          FROM judge_job
                          WHERE contest_id = ?
                            AND status = ?
                          GROUP BY inserter, problem_id) fa
                            ON j.inserter = fa.inserter AND j.problem_id = fa.problem_id
               LEFT JOIN judge_job ac ON ac.id = fa.ac_id
      WHERE j.contest_id = ?
      GROUP BY j.inserter, j.problem_id) AS flat
         LEFT JOIN user as u ON flat.inserter = u.id
GROUP BY inserter
`
		rows, err = d.db.WithContext(ctx).
			Raw(execSql, id, foundationjudge.JudgeStatusAC, id).
			Rows()
	} else {
		execSql = `
SELECT inserter,
       u.username AS username,
       u.nickname AS nickname,
       JSON_ARRAYAGG(
               JSON_OBJECT(
                       'id', problem_id,
                       'attempt', count_before,
                       'lock', count_lock,
                       'ac', ac
               )
       )          AS problems
FROM (SELECT j.inserter,
             j.problem_id,
             SUM((fa.ac_id is null AND j.insert_time < ?) OR j.id < fa.ac_id) AS count_before,
             SUM(fa.ac_id is not null AND j.insert_time >= ?)                 AS count_lock,
             ac.insert_time                                                                       AS ac
      FROM judge_job j
               LEFT JOIN (SELECT inserter, problem_id, MIN(id) AS ac_id
                          FROM judge_job
                          WHERE contest_id = ?
                            AND status = ?
                            AND insert_time < ?
                          GROUP BY inserter, problem_id) fa
                            ON j.inserter = fa.inserter AND j.problem_id = fa.problem_id
               LEFT JOIN judge_job ac ON ac.id = fa.ac_id
      WHERE j.contest_id = ?
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

func (d *JudgeJobDao) GetJudgeJobCountNotFinish(ctx context.Context) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Where("status <= ?", foundationjudge.JudgeStatusRunning).
		Count(&count).Error
	if err != nil {
		return 0, metaerror.Wrap(err, "failed to count judge jobs")
	}
	return int(count), nil
}

func (d *JudgeJobDao) ForeachContestAcCodes(
	ctx context.Context,
	contestId int,
	handleCode func(judgeId int, code string, problemId string, createTime time.Time, authorId int) error,
) error {
	rows, err := d.db.WithContext(ctx).
		Table("judge_job").
		Select("id, code, problem_id, insert_time, inserter").
		Where("contest_id = ? AND status = ?", contestId, foundationjudge.JudgeStatusAC).
		Rows()
	if err != nil {
		return fmt.Errorf("failed to query submissions: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close rows"))
		}
	}()
	for rows.Next() {
		var (
			id         int
			code       string
			problemId  string
			insertTime time.Time
			inserter   int
		)
		if err := rows.Scan(&id, &code, &problemId, &insertTime, &inserter); err != nil {
			return metaerror.Wrap(err, "failed to scan row")
		}
		if err := handleCode(id, code, problemId, insertTime, inserter); err != nil {
			return metaerror.Wrap(err, "failed to handle code")
		}
	}
	return nil
}

// RequestJudgeJobListPendingJudge 获取待评测的 JudgeJob 列表，优先取最小的
func (d *JudgeJobDao) RequestJudgeJobListPendingJudge(
	ctx context.Context,
	maxCount int,
	judger string,
) ([]*foundationmodel.JudgeJob, error) {
	now := time.Now()
	var jobs []*foundationmodel.JudgeJob
	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. SELECT ... FOR UPDATE SKIP LOCKED
			var jobIds []struct {
				Id int `gorm:"column:id"`
			}
			if err := tx.Raw(
				`SELECT id FROM judge_job  WHERE status IN (?, ?) ORDER BY status ,id  LIMIT ? FOR UPDATE SKIP LOCKED`,
				foundationjudge.JudgeStatusInit,
				foundationjudge.JudgeStatusRejudge,
				maxCount,
			).Scan(&jobIds).Error; err != nil {
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
			// 2. 执行 UPDATE
			if err := tx.Model(&foundationmodel.JudgeJob{}).
				Where("id IN ?", ids).
				Updates(
					map[string]interface{}{
						"status":     foundationjudge.JudgeStatusQueuing,
						"judger":     judger,
						"judge_time": now,
					},
				).Error; err != nil {
				return err
			}
			// 3. SELECT 完整任务信息返回
			return tx.Where("id IN ?", ids).Find(&jobs).Error
		},
	)
	if err != nil {
		return nil, err
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

func (d *JudgeJobDao) RejudgeSearch(
	ctx context.Context,
	problemId int,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
) error {
	const batchSize = 1000

	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. 加锁查找 judge_job（防止并发修改）
			var jobs []struct {
				ID        int                         `gorm:"column:id"`
				ProblemId int                         `gorm:"column:problem_id"`
				Inserter  int                         `gorm:"column:inserter"`
				Status    foundationjudge.JudgeStatus `gorm:"column:status"`
			}
			db := tx.Table("judge_job AS j").
				Select("j.id, j.problem_id, j.inserter, j.status").
				Where(
					"EXISTS (?)",
					tx.Table("problem_local AS pr").
						Select("1").
						Where("pr.problem_id = j.problem_id"),
				)

			if problemId > 0 {
				db = db.Where("j.problem_id = ?", problemId)
			}
			if foundationjudge.IsValidJudgeLanguage(int(language)) {
				db = db.Where("j.language = ?", language)
			}
			if foundationjudge.IsValidJudgeStatus(int(status)) {
				db = db.Where("j.status = ?", status)
			}

			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
				Find(&jobs).Error; err != nil {
				return metaerror.Wrap(err, "failed to find judge_jobs")
			}
			if len(jobs) == 0 {
				return nil // 没有任务可处理
			}

			// 2. 计算更新偏移
			problemAcceptDelta := map[int]int{}
			userAcceptDelta := map[int]int{}
			ids := make([]int, 0, len(jobs))
			for _, job := range jobs {
				ids = append(ids, job.ID)
				if job.Status == foundationjudge.JudgeStatusAC {
					problemAcceptDelta[job.ProblemId]--
					userAcceptDelta[job.Inserter]--
				}
			}

			// 3. 分批更新 judge_job 及删除相关数据
			updateMap := map[string]interface{}{
				"status":       foundationjudge.JudgeStatusRejudge,
				"score":        nil,
				"time":         nil,
				"memory":       nil,
				"task_current": nil,
				"task_total":   nil,
				"judger":       nil,
				"judge_time":   nil,
			}

			for start := 0; start < len(ids); start += batchSize {
				end := start + batchSize
				if end > len(ids) {
					end = len(ids)
				}
				batch := ids[start:end]

				if err := tx.Table("judge_job").
					Where("id IN ?", batch).
					Updates(updateMap).Error; err != nil {
					return metaerror.Wrap(err, "failed to update judge_job batch")
				}

				if err := tx.Table("judge_job_compile").
					Where("id IN ?", batch).
					Delete(nil).Error; err != nil {
					return metaerror.Wrap(err, "failed to delete judge_job_compile batch")
				}

				if err := tx.Table("judge_task").
					Where("id IN ?", batch).
					Delete(nil).Error; err != nil {
					return metaerror.Wrap(err, "failed to delete judge_task batch")
				}
			}

			// 4. 更新 problem 和 user 的 accept 计数
			for pid, delta := range problemAcceptDelta {
				if delta != 0 {
					if err := tx.Table("problem").
						Where("id = ?", pid).
						Update("accept", gorm.Expr("accept + ?", delta)).Error; err != nil {
						return metaerror.Wrap(err, "failed to update problem accept for problem_id=%d", pid)
					}
				}
			}
			for uid, delta := range userAcceptDelta {
				if delta != 0 {
					if err := tx.Table("user").
						Where("id = ?", uid).
						Update("accept", gorm.Expr("accept + ?", delta)).Error; err != nil {
						return metaerror.Wrap(err, "failed to update user accept for user_id=%d", uid)
					}
				}
			}
			return nil
		},
	)
}

func (d *JudgeJobDao) RejudgeRecently(ctx context.Context) error {
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. 加锁查找 judge_job（防止并发修改）
			var jobs []struct {
				ID        int                         `gorm:"column:id"`
				ProblemId int                         `gorm:"column:problem_id"`
				Inserter  int                         `gorm:"column:inserter"`
				Status    foundationjudge.JudgeStatus `gorm:"column:status"`
			}
			db := tx.Table("judge_job AS j").
				Select("j.id, j.problem_id, j.inserter, j.status").
				Where(
					"EXISTS (?)",
					tx.Table("problem_local AS pr").
						Select("1").
						Where("pr.problem_id = j.problem_id"),
				).Order("id desc").Limit(100)

			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
				Find(&jobs).Error; err != nil {
				return metaerror.Wrap(err, "failed to find judge_jobs")
			}
			if len(jobs) == 0 {
				return nil // 没有任务可处理
			}

			// 2. 计算更新偏移
			problemAcceptDelta := map[int]int{}
			userAcceptDelta := map[int]int{}
			ids := make([]int, 0, len(jobs))
			for _, job := range jobs {
				ids = append(ids, job.ID)
				if job.Status == foundationjudge.JudgeStatusAC {
					problemAcceptDelta[job.ProblemId]--
					userAcceptDelta[job.Inserter]--
				}
			}

			// 3. 更新 judge_job
			updateMap := map[string]interface{}{
				"status":       foundationjudge.JudgeStatusRejudge,
				"score":        nil,
				"time":         nil,
				"memory":       nil,
				"task_current": nil,
				"task_total":   nil,
				"judger":       nil,
				"judge_time":   nil,
			}

			if err := tx.Table("judge_job").
				Where("id IN ?", ids).
				Updates(updateMap).Error; err != nil {
				return metaerror.Wrap(err, "failed to update judge_job")
			}

			// 4. 删除 judge_job_compile 中对应记录
			if err := tx.Table("judge_job_compile").
				Where("id IN ?", ids).
				Delete(nil).Error; err != nil {
				return metaerror.Wrap(err, "failed to delete judge_job_compile")
			}

			if err := tx.Table("judge_task").
				Where("id IN ?", ids).
				Delete(nil).Error; err != nil {
				return metaerror.Wrap(err, "failed to delete judge_task")
			}

			for pid, delta := range problemAcceptDelta {
				if delta != 0 {
					if err := tx.Table("problem").
						Where("id = ?", pid).
						Update("accept", gorm.Expr("accept + ?", delta)).Error; err != nil {
						return metaerror.Wrap(err, "failed to update problem accept for problem_id=%d", pid)
					}
				}
			}
			for uid, delta := range userAcceptDelta {
				if delta != 0 {
					if err := tx.Table("user").
						Where("id = ?", uid).
						Update("accept", gorm.Expr("accept + ?", delta)).Error; err != nil {
						return metaerror.Wrap(err, "failed to update user accept for user_id=%d", uid)
					}
				}
			}
			return nil
		},
	)
}

func (d *JudgeJobDao) RejudgeAll(ctx context.Context) error {
	const batchSize = 1000
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. 加锁查找 judge_job（防止并发修改）
			var jobs []struct {
				ID        int                         `gorm:"column:id"`
				ProblemId int                         `gorm:"column:problem_id"`
				Inserter  int                         `gorm:"column:inserter"`
				Status    foundationjudge.JudgeStatus `gorm:"column:status"`
			}
			db := tx.Table("judge_job AS j").
				Select("j.id, j.problem_id, j.inserter, j.status").
				Where(
					"EXISTS (?)",
					tx.Table("problem_local AS pr").
						Select("1").
						Where("pr.problem_id = j.problem_id"),
				)

			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
				Find(&jobs).Error; err != nil {
				return metaerror.Wrap(err, "failed to find judge_jobs")
			}
			if len(jobs) == 0 {
				return nil // 没有任务可处理
			}

			// 2. 计算更新偏移
			problemAcceptDelta := map[int]int{}
			userAcceptDelta := map[int]int{}
			ids := make([]int, 0, len(jobs))
			for _, job := range jobs {
				ids = append(ids, job.ID)
				if job.Status == foundationjudge.JudgeStatusAC {
					problemAcceptDelta[job.ProblemId]--
					userAcceptDelta[job.Inserter]--
				}
			}

			// 3. 分批更新 judge_job
			updateMap := map[string]interface{}{
				"status":       foundationjudge.JudgeStatusRejudge,
				"score":        nil,
				"time":         nil,
				"memory":       nil,
				"task_current": nil,
				"task_total":   nil,
				"judger":       nil,
				"judge_time":   nil,
			}

			for start := 0; start < len(ids); start += batchSize {
				end := start + batchSize
				if end > len(ids) {
					end = len(ids)
				}
				batch := ids[start:end]

				if err := tx.Table("judge_job").
					Where("id IN ?", batch).
					Updates(updateMap).Error; err != nil {
					return metaerror.Wrap(err, "failed to update judge_job batch")
				}

				if err := tx.Table("judge_job_compile").
					Where("id IN ?", batch).
					Delete(nil).Error; err != nil {
					return metaerror.Wrap(err, "failed to delete judge_job_compile batch")
				}

				if err := tx.Table("judge_task").
					Where("id IN ?", batch).
					Delete(nil).Error; err != nil {
					return metaerror.Wrap(err, "failed to delete judge_task batch")
				}
			}

			// 4. 更新 problem 和 user 的 accept 计数
			for pid, delta := range problemAcceptDelta {
				if delta != 0 {
					if err := tx.Table("problem").
						Where("id = ?", pid).
						Update("accept", gorm.Expr("accept + ?", delta)).Error; err != nil {
						return metaerror.Wrap(err, "failed to update problem accept for problem_id=%d", pid)
					}
				}
			}
			for uid, delta := range userAcceptDelta {
				if delta != 0 {
					if err := tx.Table("user").
						Where("id = ?", uid).
						Update("accept", gorm.Expr("accept + ?", delta)).Error; err != nil {
						return metaerror.Wrap(err, "failed to update user accept for user_id=%d", uid)
					}
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
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			if judgeJob == nil {
				return metaerror.New("judgeJob is nil")
			}
			if err := tx.Create(judgeJob).Error; err != nil {
				return metaerror.Wrap(err, "insert judgeJob")
			}
			// 标记用户attempt
			if err := tx.Model(&foundationmodel.User{}).
				Where("id = ?", judgeJob.Inserter).
				UpdateColumn("attempt", gorm.Expr("attempt + ?", 1)).Error; err != nil {
				return metaerror.Wrap(err, "update user attempt count")
			}
			// 标记问题attempt
			if err := tx.Model(&foundationmodel.Problem{}).
				Where("id = ?", judgeJob.ProblemId).
				UpdateColumn("attempt", gorm.Expr("attempt + ?", 1)).Error; err != nil {
				return metaerror.Wrap(err, "update problem attempt count")
			}
			// 做个保底判断，如果是 AC 状态，更新相关的 accept 计数
			if judgeJob.Status == foundationjudge.JudgeStatusAC {
				// 更新 problem accept 计数
				if err := tx.Model(&foundationmodel.Problem{}).
					Where("id = ?", judgeJob.ProblemId).
					UpdateColumn("accept", gorm.Expr("accept + ?", 1)).Error; err != nil {
					return metaerror.Wrap(err, "update problem accept count")
				}
				// 更新 user accept 计数
				if err := tx.Model(&foundationmodel.User{}).
					Where("id = ?", judgeJob.Inserter).
					UpdateColumn("accept", gorm.Expr("accept + ?", 1)).Error; err != nil {
					return metaerror.Wrap(err, "update user accept count")
				}
			}
			return nil
		},
	)
}
