package service

import (
	"context"
	"log/slog"
	"meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"migrate/migrate"
	"sort"
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
type CodeojStatus struct {
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

type JolSolution struct {
	SolutionID int        `gorm:"column:solution_id"`
	ProblemID  int        `gorm:"column:problem_id"`
	UserId     string     `gorm:"column:user_id"`
	Time       int        `gorm:"column:time"`
	Memory     int        `gorm:"column:memory"`
	InData     time.Time  `gorm:"column:in_data"`
	Result     int        `gorm:"column:result"`
	Language   int        `gorm:"column:language"`
	ContestId  int        `gorm:"column:contest_id"`
	Num        int        `gorm:"column:num"`
	CodeLength int        `gorm:"column:code_length"`
	JudgeTime  *time.Time `gorm:"column:judge_time"`
	Judger     string     `gorm:"column:judger"`
	Code       string     `gorm:"column:code"`
	Private    int        `gorm:"column:pr"`
}

func (s *MigrateJudgeJobService) Start() error {
	ctx := context.Background()

	var judgeJobs []*foundationmodel.JudgeJob

	codeojJudgeJobs, err := s.processCodeojJudgeJob(ctx)
	if err != nil {
		return err
	}
	judgeJobs = append(judgeJobs, codeojJudgeJobs...)

	jolJudgeJobs, err := s.processJolJudgeJob(ctx)
	if err != nil {
		return err
	}
	judgeJobs = append(judgeJobs, jolJudgeJobs...)

	slog.Info("migrate judge job updates", "count", len(judgeJobs))

	sort.Slice(judgeJobs, func(i, j int) bool {
		return judgeJobs[i].ApproveTime.Before(judgeJobs[j].ApproveTime)
	})

	for _, judgeJob := range judgeJobs {
		err = foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
		if err != nil {
			return metaerror.Wrap(err, "insert judge job failed")
		}
	}

	return nil
}

func (s *MigrateJudgeJobService) processCodeojJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {

	codeojDB := metamysql.GetSubsystem().GetClient("codeoj")

	const batchSize = 1000
	offset := 0

	slog.Info("migrate judge job start", "batchSize", batchSize)

	var judgeJobs []*foundationmodel.JudgeJob

	for {
		var rows []CodeojStatus

		err := codeojDB.Table("status AS s").
			Select("s.status_id, s.problem_id, s.creator, s.language, s.insert_time, s.length, s.time, s.memory, s.result, s.score, s.judge_time, s.judger, c.code").
			Joins("LEFT JOIN status_code c ON s.status_id = c.status_id").
			Limit(batchSize).
			Offset(offset).
			Where("s.status_id > 98402").
			Scan(&rows).Error
		if err != nil {
			return nil, metaerror.Wrap(err, "query status failed")
		}

		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			userId, err := GetMigrateUserService().getUserIdByUsername(ctx, row.Creator)
			if err != nil {
				return nil, metaerror.Wrap(err, "get user id by username failed")
			}

			newProblemId := GetMigrateProblemService().GetNewProblemId(row.ProblemID)

			judgeJob := foundationmodel.NewJudgeJobBuilder().
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
		}

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return judgeJobs, nil
}

func (s *MigrateJudgeJobService) processJolJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {
	slog.Info("migrate judge job processJolJudgeJob")

	jolDB := metamysql.GetSubsystem().GetClient("jol")

	const batchSize = 1000
	offset := 0

	var judgeJobs []*foundationmodel.JudgeJob

	for {
		var rows []JolSolution

		err := jolDB.Table("solution AS s").
			Select("s.solution_id, s.problem_id, s.user_id, s.language, s.contest_id, s.num, s.code_length, s.judge_time, s.judger, c.code, pr.pr").
			Joins("LEFT JOIN source_code c ON s.solution_id = c.solution_id").
			Joins("LEFT JOIN source_code_pr pr ON s.solution_id = pr.solution_id").
			Limit(batchSize).
			Offset(offset).
			Scan(&rows).Error
		if err != nil {
			return nil, metaerror.Wrap(err, "query status failed")
		}

		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			userId, err := GetMigrateUserService().getUserIdByUsername(ctx, row.UserId)
			if err != nil {
				return nil, metaerror.Wrap(err, "get user id by username failed")
			}

			newProblemId := GetMigrateProblemService().GetNewProblemId(row.ProblemID)

			// TODO 补充contest信息

			judgeJob := foundationmodel.NewJudgeJobBuilder().
				ProblemId(newProblemId).
				Author(userId).
				ApproveTime(row.InData).
				Language(migrate.GetJudgeLanguageByCodeOJ(row.Language)).
				Code(row.Code).
				CodeLength(row.CodeLength).
				Status(migrate.GetJudgeStatusByCodeOJ(row.Result)).
				JudgeTime(row.JudgeTime).
				Judger(row.Judger).
				Build()

			judgeJobs = append(judgeJobs, judgeJob)
		}

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return judgeJobs, nil
}
