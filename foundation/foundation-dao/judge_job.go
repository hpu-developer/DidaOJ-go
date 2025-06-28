package foundationdao

import (
	"context"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
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

func (d *JudgeJobDao) GetProblemAttemptStatus(
	ctx context.Context, problemIds []int, authorId int,
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
		default:
			statusMap[r.ProblemId] = foundationenum.ProblemAttemptStatusNone
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
