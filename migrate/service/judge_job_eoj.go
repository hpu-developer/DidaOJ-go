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

type MigrateJudgeJobEojService struct{}

var singletonMigrateJudgeJobEojService = singleton.Singleton[MigrateJudgeJobEojService]{}

func GetMigrateJudgeJobEojService() *MigrateJudgeJobEojService {
	return singletonMigrateJudgeJobEojService.GetInstance(
		func() *MigrateJudgeJobEojService {
			return &MigrateJudgeJobEojService{}
		},
	)
}

type EojSubmission struct {
	Id         int       `gorm:"column:id"`
	Lang       string    `gorm:"column:lang"`
	Code       string    `gorm:"column:code"`
	CreateTime time.Time `gorm:"column:create_time"`
	CodeLength int       `gorm:"column:code_length"`
	Visible    int       `gorm:"column:visible"`
	AuthorId   int       `gorm:"column:author_id"`
	ContestId  int       `gorm:"column:contest_id"`
	ProblemId  int       `gorm:"column:problem_id"`
}

func (EojSubmission) TableName() string {
	return "submission_submission"
}

func (s *MigrateJudgeJobEojService) Start() error {
	ctx := context.Background()

	var judgeJobs []*foundationmodel.JudgeJob

	eojJudgeJobs, err := s.processEojJudgeJob(ctx)
	if err != nil {
		return err
	}
	judgeJobs = append(judgeJobs, eojJudgeJobs...)

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

func (s *MigrateJudgeJobEojService) processEojJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {

	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	const batchSize = 1000
	offset := 0

	slog.Info("migrate judge job start", "batchSize", batchSize)

	var judgeJobs []*foundationmodel.JudgeJob

	for {
		var judgeJobModels []EojSubmission
		if err := eojDb.
			Select("id, lang, code, create_time, code_length, visible, author_id, contest_id, problem_id").
			Model(&EojSubmission{}).
			Order("id ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&judgeJobModels).Error; err != nil {
			return nil, metaerror.Wrap(err, "query eoj problem managers failed")
		}
		if len(judgeJobModels) == 0 {
			break
		}

		for _, judgeJobModel := range judgeJobModels {
			language := migratetype.GetJudgeLanguageByEOJ(judgeJobModel.Lang)
			if language == foundationjudge.JudgeLanguageUnknown {
				continue
			}
			newProblemId, err := migratedao.GetMigrateMarkDao().GetMark(
				ctx,
				"eoj-problem",
				strconv.Itoa(judgeJobModel.ProblemId),
			)
			if err != nil {
				return nil, err
			}
			if newProblemId == nil {
				continue
			}
			var newContestId int
			if judgeJobModel.ContestId > 0 {
				newContestIdStr, err := migratedao.GetMigrateMarkDao().GetMark(
					ctx,
					"eoj-contest",
					strconv.Itoa(judgeJobModel.ContestId),
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "get eoj contest mark failed")
				}
				if newContestIdStr != nil {
					newContestId, err = strconv.Atoi(*newContestIdStr)
					if err != nil {
						return nil, metaerror.Wrap(err, "convert eoj contest id to int failed")
					}
				}
			}
			userId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(judgeJobModel.AuthorId))
			if err != nil {
				return nil, metaerror.Wrap(err, "get eoj user mark failed")
			}
			realUserId, err := strconv.Atoi(*userId)
			if err != nil {
				return nil, metaerror.Wrap(err, "convert eoj user id to int failed")
			}
			judgeJob := foundationmodel.NewJudgeJobBuilder().
				ProblemId(*newProblemId).
				ContestId(newContestId).
				AuthorId(realUserId).
				ApproveTime(judgeJobModel.CreateTime).
				Language(language).
				Code(judgeJobModel.Code).
				CodeLength(judgeJobModel.CodeLength).
				Status(foundationjudge.JudgeStatusUnknown).
				Score(0).
				Private(judgeJobModel.Visible <= 0).
				Build()

			judgeJobs = append(judgeJobs, judgeJob)
		}

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return judgeJobs, nil
}

func (s *MigrateJudgeJobEojService) processDidaOjJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {
	return migratedao.GetJudgeJobDao().GetJudgeJobList(ctx)
}
