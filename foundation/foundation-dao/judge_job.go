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
