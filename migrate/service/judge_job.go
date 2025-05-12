package service

import (
	"context"
	"log/slog"
	"meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"migrate/migrate"
	"time"

	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
)

type MigrateJudgeJobService struct{}

var singletonMigrateJudgeJobService = singleton.Singleton[MigrateJudgeJobService]{}

func GetMigrateJudgeJobService() *MigrateJudgeJobService {
	return singletonMigrateJudgeJobService.GetInstance(
		func() *MigrateJudgeJobService {
			return &MigrateJudgeJobService{}
		},
	)
}

// GORM 结构体映射
type Status struct {
	StatusID   int        `gorm:"column:status_id"`
	ProblemID  int        `gorm:"column:problem_id"`
	Creator    string     `gorm:"column:creator"`
	Language   int        `gorm:"column:language"`
	InsertTime time.Time  `gorm:"column:insert_time"`
	Length     int        `gorm:"column:length"`
	Time       int        `gorm:"column:time"`
	Memory     int        `gorm:"column:memory"`
	Result     int        `gorm:"column:result"`
	Score      int        `gorm:"column:score"`
	JudgeTime  *time.Time `gorm:"column:judge_time"` // 允许为空
	Judger     string     `gorm:"column:judger"`
	Code       string     `gorm:"column:code"`
}

func (s *MigrateJudgeJobService) Start() error {
	ctx := context.Background()

	codeojDB := metamysql.GetSubsystem().GetClient("codeoj")

	const batchSize = 1000
	offset := 0
	mongoStatusId := 1

	slog.Info("migrate judge job start", "batchSize", batchSize)

	usernameToUserId := make(map[string]int)

	for {
		var judgeJobs []*foundationmodel.JudgeJob
		var rows []Status

		err := codeojDB.Table("status AS s").
			Select("s.status_id, s.problem_id, s.creator, s.language, s.insert_time, s.length, s.time, s.memory, s.result, s.score, s.judge_time, s.judger, c.code").
			Joins("LEFT JOIN status_code c ON s.status_id = c.status_id").
			Limit(batchSize).
			Offset(offset).
			Scan(&rows).Error
		if err != nil {
			return metaerror.Wrap(err, "query status failed")
		}

		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			newProblemId := GetMigrateProblemService().GetNewProblemId(row.ProblemID)

			userId, ok := usernameToUserId[row.Creator]
			if !ok {
				userId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, row.Creator)
				if err != nil {
					return err
				}
				usernameToUserId[row.Creator] = userId
			}

			judgeJob := foundationmodel.NewJudgeJobBuilder().
				Id(mongoStatusId).
				ProblemId(newProblemId).
				Author(userId).
				ApproveTime(row.InsertTime).
				Language(migrate.GetJudgeLanguageByCodeOJ(row.Language)).
				Code(row.Code).
				CodeLength(row.Length).
				Status(migrate.GetJudgeStatusByCodeOJ(row.Result)).
				Score(row.Score).
				JudgeTime(row.JudgeTime).
				Judger(row.Judger).
				Build()

			judgeJobs = append(judgeJobs, judgeJob)
			mongoStatusId++
		}

		err = foundationdao.GetJudgeJobDao().UpdateJudgeJobs(ctx, judgeJobs)
		if err != nil {
			return err
		}

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return nil
}
