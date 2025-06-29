package foundationdao

import (
	"context"
	"database/sql"
	"errors"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"gorm.io/gorm"
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
		)
	} else {
		selectFields = []string{
			"j.*",
			"u.username AS inserter_username",
			"u.nickname AS inserter_nickname",
			"judger.name AS judger_name",
		}
	}
	err := d.db.WithContext(ctx).Table("judge_job AS j").
		Select(strings.Join(selectFields, ", ")).
		Joins("LEFT JOIN users AS u ON u.id = j.inserter").
		Joins("LEFT JOIN judger AS judger ON judger.key = j.judger").
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
	db := d.db.WithContext(ctx).Table("judge_job AS j").
		Select(
			`
			j.id, j.approve_time, j.language, j.score, j.status,
			j.time, j.memory, j.problem_id, j.author_id, j.code_length,
			u.username AS inserter_username, u.nickname AS inserter_nickname,
			cp.index AS contest_problem_index
		`,
		).
		Joins("LEFT JOIN users AS u ON u.id = j.author_id")
	if contestId > 0 {
		db = db.Joins(
			`
			LEFT JOIN contest_problem AS cp
			ON cp.id = j.contest_id AND cp.problem_id = j.problem_id
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
		db = db.Where("j.author_id = ?", searchUserId)
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
		Where("author_id = ?", userId).
		Pluck("problem_id", &problemIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to get distinct problem_ids")
	}
	return problemIds, nil
}

func (d *JudgeJobDao) GetAcUserIds(db *gorm.DB, problemId int, limit int) ([]int, error) {
	var acUserIds []int
	subDb := db.Model(&foundationmodel.JudgeJob{}).
		Select("DISTINCT author_id").
		Where("status = ?", foundationjudge.JudgeStatusAC)
	if problemId > 0 {
		subDb = subDb.Where("problem_id = ?", problemId)
	}
	subDb = subDb.Limit(1000)
	if err := subDb.Pluck("author_id", &acUserIds).Error; err != nil {
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
		Where("jj.approve_time IS NOT NULL").
		Where("jj.author_id IN ?", acUserIDs).
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
		db = db.Where("approve_time >= ?", *approveStartTime)
	}
	if approveEndTime != nil {
		db = db.Where("approve_time < ?", *approveEndTime)
	}
	subQuery := db.
		Select("author_id AS id, COUNT(DISTINCT problem_id) AS problem_count").
		Group("author_id")
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
                       'ac', ac
               )
       )          AS problems
FROM (SELECT j.inserter,
             j.problem_id,
             COUNT(*)       AS count,
             ac.insert_time AS ac
      FROM judge_job j
               LEFT JOIN (SELECT inserter, problem_id, MIN(id) AS ac_id
                          FROM judge_job
                          WHERE contest_id = 10
                            AND status = 6
                          GROUP BY inserter, problem_id) fa ON j.inserter = fa.inserter AND j.problem_id = fa.problem_id
               LEFT JOIN judge_job ac ON ac.id = fa.ac_id
      WHERE j.contest_id = 10
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
                       'ac', ac
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

		rows, err = d.db.WithContext(ctx).Raw(
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
		return nil, err
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
		err := rows.Scan(&rank)
		if err != nil {
			return nil, err
		}
		for _, problem := range rank.Problems {
			problem.Index = problemMap[problem.Id]
			problem.Id = 0
		}
		ranks = append(ranks, &rank)
	}
	return ranks, nil
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
