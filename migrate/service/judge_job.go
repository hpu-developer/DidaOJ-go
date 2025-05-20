package service

import (
	"context"
	"fmt"
	foundationjudge "foundation/foundation-judge"
	"log/slog"
	"meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"migrate/type"
	"sort"
	"strconv"
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
	InDate     time.Time  `gorm:"column:in_date"`
	Result     int        `gorm:"column:result"`
	Language   int        `gorm:"column:language"`
	ContestId  int        `gorm:"column:contest_id"`
	Num        int        `gorm:"column:num"`
	CodeLength int        `gorm:"column:code_length"`
	JudgeTime  *time.Time `gorm:"column:judgetime"`
	Judger     string     `gorm:"column:judger"`
	Code       string     `gorm:"column:source"`
	Private    int        `gorm:"column:pr"`
}

type VhojSubmission struct {
	Id                int       `gorm:"column:C_ID"`
	Time              int       `gorm:"column:C_TIME"`
	Memory            int       `gorm:"column:C_MEMORY"`
	SubTime           time.Time `gorm:"column:C_SUBTIME"`
	ProblemId         int       `gorm:"column:C_PROBLEM_ID"`
	ContestId         int       `gorm:"column:C_CONTEST_ID"`
	ContestNum        string    `gorm:"column:C_CONTEST_NUM"`
	Source            string    `gorm:"column:C_SOURCE"`
	IsOpen            bool      `gorm:"column:C_ISOPEN"`
	UserId            int       `gorm:"column:C_USER_ID"`
	Username          string    `gorm:"column:C_USERNAME"`
	OriginOj          string    `gorm:"column:C_ORIGIN_OJ"`
	OriginProb        string    `gorm:"column:C_ORIGIN_PROB"`
	IsPrivate         bool      `gorm:"column:C_IS_PRIVATE"`
	AdditionalInfo    string    `gorm:"column:C_ADDITIONAL_INFO"`
	RealRunId         string    `gorm:"column:C_REAL_RUNID"`
	RemoteAccountId   string    `gorm:"column:C_REMOTE_ACCOUNT_ID"`
	QueryCount        int       `gorm:"column:C_QUERY_COUNT"`
	StatusUpdateTime  time.Time `gorm:"column:C_STATUS_UPDATE_TIME"`
	RemoteSubmitTime  time.Time `gorm:"column:C_REMOTE_SUBMIT_TIME"`
	Status            string    `gorm:"column:C_STATUS"` // RemoteStatus
	StatusCanonical   string    `gorm:"column:C_STATUS_CANONICAL"`
	SourceLength      int       `gorm:"column:C_SOURCE_LENGTH"`
	Language          string    `gorm:"column:C_LANGUAGE"`
	DispLanguage      string    `gorm:"column:C_DISP_LANGUAGE"`
	LanguageCanonical string    `gorm:"column:C_LANGUAGE_CANONICAL"`
}

func (VhojSubmission) TableName() string {
	return "t_submission"
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

	vhojJudgeJobs, err := s.processVhojJudgeJob(ctx)
	if err != nil {
		return err
	}
	judgeJobs = append(judgeJobs, vhojJudgeJobs...)

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

			if newProblemId == "-1" {
				return nil, metaerror.New("problem id not found", "oldProblemId", row.ProblemID)
			}

			judgeJob := foundationmodel.NewJudgeJobBuilder().
				ProblemId(newProblemId).
				AuthorId(userId).
				ApproveTime(row.InsertTime).
				Language(migratetype.GetJudgeLanguageByCodeOJ(row.Language)).
				Code(row.Code).
				CodeLength(row.Length).
				Status(migratetype.GetJudgeStatusByCodeOJ(row.Result)).
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
			Select("s.solution_id, s.problem_id, s.user_id, s.time, s.memory, s.in_date, s.result, s.language, s.contest_id, s.num, s.code_length, s.judgetime, s.judger, c.source, pr.pr").
			Joins("LEFT JOIN source_code c ON s.solution_id = c.solution_id").
			Joins("LEFT JOIN source_code_pr pr ON s.solution_id = pr.solution_id").
			Where("s.problem_id > 0").
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

			if newProblemId == "-1" {
				return nil, metaerror.New("problem id not found", "oldProblemId:%d", row.ProblemID)
			}

			judgeJob := foundationmodel.NewJudgeJobBuilder().
				ProblemId(newProblemId).
				AuthorId(userId).
				ApproveTime(row.InDate).
				Language(migratetype.GetJudgeLanguageByCodeOJ(row.Language)).
				Code(row.Code).
				CodeLength(row.CodeLength).
				Status(migratetype.GetJudgeStatusByCodeOJ(row.Result)).
				JudgeTime(row.JudgeTime).
				Judger(row.Judger).
				Build()

			if row.ContestId > 0 {
				judgeJob.ContestId = GetMigrateContestService().GetNewContestIdByJol(row.ContestId)
			}

			judgeJobs = append(judgeJobs, judgeJob)
		}

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return judgeJobs, nil
}

func (s *MigrateJudgeJobService) processVhojJudgeJob(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {
	slog.Info("migrate judge job processVhojJudgeJob")

	vhojDB := metamysql.GetSubsystem().GetClient("vhoj")

	var judgeJobs []*foundationmodel.JudgeJob

	var rows []VhojSubmission
	if err := vhojDB.Order("C_ID ASC").Find(&rows).Error; err != nil {
		return nil, metaerror.Wrap(err, "query judge job failed")
	}
	for _, row := range rows {
		vhojUsername, err := GetMigrateUserService().getUsernameByVhojId(row.UserId)
		if err != nil {
			return nil, metaerror.Wrap(err, "get username by vhoj id failed")
		}
		userId, err := GetMigrateUserService().getUserIdByUsername(ctx, vhojUsername)
		if err != nil {
			return nil, metaerror.Wrap(err, "get user id by username failed")
		}

		var newProblemId string
		if row.OriginOj == "HPU" {
			hpuId, err := strconv.Atoi(row.OriginProb)
			if err != nil {
				return nil, metaerror.Wrap(err, "parse hpu id failed")
			}
			row.OriginOj = ""
			row.OriginProb = ""
			row.RealRunId = ""
			row.RemoteAccountId = ""
			row.Language = ""
			newProblemId = GetMigrateProblemService().GetNewProblemId(hpuId)
		} else {
			newProblemId = fmt.Sprintf("%s-%s", row.OriginOj, row.OriginProb)
		}

		if newProblemId == "-1" {
			return nil, metaerror.New("problem id not found", "oldProblemId", row.OriginProb)
		}

		judgeJob := foundationmodel.NewJudgeJobBuilder().
			ProblemId(newProblemId).
			Time(row.Time).
			Memory(row.Memory).
			ApproveTime(row.SubTime).
			Code(row.Source).
			CodeLength(len(row.Source)).
			Private(!row.IsOpen).
			AuthorId(userId).
			OriginOj(&row.OriginOj).
			OriginProblemId(&row.OriginProb).
			CompileMessage(&row.AdditionalInfo).
			RemoteJudgeId(&row.RealRunId).
			RemoteAccountId(&row.RemoteAccountId).
			Language(migratetype.GetJudgeLanguageByVhoj(row.LanguageCanonical)).
			Status(migratetype.GetJudgeStatusByVhoj(row.StatusCanonical)).
			JudgeTime(&row.RemoteSubmitTime).
			Judger("didaoj").
			Build()

		if judgeJob.Status == foundationjudge.JudgeStatusAC {
			judgeJob.Score = 100
		}

		if row.ContestId > 0 {
			judgeJob.ContestId = GetMigrateContestService().GetNewContestIdByVhoj(row.ContestId)
		}

		judgeJobs = append(judgeJobs, judgeJob)
	}

	return judgeJobs, nil
}
