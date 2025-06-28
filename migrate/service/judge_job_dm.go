package service

import (
	"context"
	foundationdao "foundation/foundation-dao-mongo"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model-mongo"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	migratedao "migrate/dao"
	migratetype "migrate/type"
	"sort"
	"strconv"
	"time"
)

type MigrateJudgeJobDmojService struct{}

var singletonMigrateJudgeJobDmojService = singleton.Singleton[MigrateJudgeJobDmojService]{}

func GetMigrateJudgeJobDmojService() *MigrateJudgeJobDmojService {
	return singletonMigrateJudgeJobDmojService.GetInstance(
		func() *MigrateJudgeJobDmojService {
			return &MigrateJudgeJobDmojService{}
		},
	)
}

type DmojSubmission struct {
	SubmitId    int       `gorm:"column:submit_id"`
	Pid         int       `gorm:"column:pid"` // Problem ID
	Uid         string    `gorm:"column:uid"` // User ID
	SubmitTime  time.Time `gorm:"column:submit_time"`
	Length      int       `gorm:"column:length"`
	Code        string    `gorm:"column:code"`
	Language    string    `gorm:"column:language"`
	GmtCreate   time.Time `gorm:"column:gmt_create"`
	GmtModified time.Time `gorm:"column:gmt_modified"`
}

func (DmojSubmission) TableName() string {
	return "judge"
}

func (s *MigrateJudgeJobDmojService) Start() error {
	ctx := context.Background()

	var judgeJobs []*foundationmodel.JudgeJob

	dmojJudgeJobs, err := s.processDmojJudgeJob(ctx)
	if err != nil {
		return err
	}
	judgeJobs = append(judgeJobs, dmojJudgeJobs...)

	didaojJudgeJobs, err := s.processDidaOjJudgeJob(ctx)
	if err != nil {
		return err
	}
	judgeJobs = append(judgeJobs, didaojJudgeJobs...)

	slog.Info("migrate judge job updates", "count", len(judgeJobs))

	sort.Slice(
		judgeJobs, func(i, j int) bool {
			return judgeJobs[i].ApproveTime.Before(judgeJobs[j].ApproveTime)
		},
	)

	for _, judgeJob := range judgeJobs {
		err = foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
		if err != nil {
			return metaerror.Wrap(err, "insert judge job failed")
		}
	}

	return nil
}

func (s *MigrateJudgeJobDmojService) processDmojJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {

	dmojDb := metamysql.GetSubsystem().GetClient("dmoj")

	const batchSize = 1000
	offset := 0

	slog.Info("migrate judge job start", "batchSize", batchSize)

	var judgeJobs []*foundationmodel.JudgeJob

	for {
		var judgeJobModels []DmojSubmission
		if err := dmojDb.
			Model(&DmojSubmission{}).
			Order("submit_id ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&judgeJobModels).Error; err != nil {
			return nil, metaerror.Wrap(err, "query dmoj problem managers failed")
		}
		if len(judgeJobModels) == 0 {
			break
		}

		for _, judgeJobModel := range judgeJobModels {
			language := migratetype.GetJudgeLanguageByEOJ(judgeJobModel.Language)
			if language == foundationjudge.JudgeLanguageUnknown {
				continue
			}
			newProblemId, err := migratedao.GetMigrateMarkDao().GetMark(
				ctx,
				"dmoj-problem",
				strconv.Itoa(judgeJobModel.Pid),
			)
			if err != nil {
				return nil, err
			}
			if newProblemId == nil {
				continue
			}
			userId, err := migratedao.GetMigrateMarkDao().GetMark(
				ctx,
				"dmoj-user",
				judgeJobModel.Uid,
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "get dmoj user mark failed")
			}
			realUserId, err := strconv.Atoi(*userId)
			if err != nil {
				return nil, metaerror.Wrap(err, "convert dmoj user id to int failed")
			}
			judgeJob := foundationmodel.NewJudgeJobBuilder().
				ProblemId(*newProblemId).
				AuthorId(realUserId).
				ApproveTime(judgeJobModel.SubmitTime).
				Language(language).
				Code(judgeJobModel.Code).
				CodeLength(judgeJobModel.Length).
				Status(foundationjudge.JudgeStatusUnknown).
				Score(0).
				Build()

			judgeJobs = append(judgeJobs, judgeJob)
		}

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return judgeJobs, nil
}

func (s *MigrateJudgeJobDmojService) processDidaOjJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {
	return migratedao.GetJudgeJobDao().GetJudgeJobList(ctx)
}
