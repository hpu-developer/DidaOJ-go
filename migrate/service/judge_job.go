package service

import (
	"context"
	"database/sql"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"migrate/migrate"
	"time"
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

func (s *MigrateJudgeJobService) Start() error {

	ctx := context.Background()

	codeojMysqlClient := metamysql.GetSubsystem().GetClient("codeoj")

	// Problem 定义
	type Problem struct {
		ProblemID   int
		Title       sql.NullString
		Description sql.NullString
		Hint        sql.NullString
		Source      sql.NullString
		Creator     sql.NullString
		Privilege   sql.NullInt64
		TimeLimit   sql.NullInt64
		MemoryLimit sql.NullInt64
		JudgeType   sql.NullInt64
		Accept      sql.NullInt64
		Attempt     sql.NullInt64
		InsertTime  sql.NullTime
		UpdateTime  sql.NullTime
	}
	type Tag struct {
		Name string
	}

	const batchSize = 1000
	offset := 0

	mongoStatusId := 1

	slog.Info("migrate judge job start", "batchSize", batchSize)

	usernameToUserId := make(map[string]int)

	for {
		var judgeJobs []*foundationmodel.JudgeJob

		rows, err := codeojMysqlClient.Query(`
		SELECT s.status_id, s.problem_id, s.creator, s.language, s.insert_time, 
		       s.length, s.time, s.memory, s.result, s.score, s.judge_time, s.judger, c.code
		FROM status s
		LEFT JOIN status_code c ON s.status_id = c.status_id
			LIMIT ? OFFSET ?
		`, batchSize, offset)
		if err != nil {
			return metaerror.Wrap(err, "query status failed")
		}

		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				metapanic.ProcessError(err)
			}
		}(rows)

		hasData := false

		for rows.Next() {

			hasData = true

			var (
				statusId   int
				problemId  int
				creator    string
				language   int
				insertTime time.Time
				length     int
				useTime    int
				useMemory  int
				result     int
				score      int
				judgeTime  sql.NullTime
				judger     string
				code       sql.NullString
			)
			err := rows.Scan(&statusId, &problemId, &creator, &language, &insertTime,
				&length, &useTime, &useMemory, &result, &score, &judgeTime, &judger, &code)
			if err != nil {
				return metaerror.Wrap(err, "scan status failed")
			}

			newProblemId := GetMigrateProblemService().GetNewProblemId(problemId)

			userId, ok := usernameToUserId[creator]
			if !ok {
				userId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, creator)
				if err != nil {
					return err
				}
				usernameToUserId[creator] = userId
			}

			judgeJob := foundationmodel.NewJudgeJobBuilder().
				Id(mongoStatusId).
				ProblemId(newProblemId).
				Author(userId).
				ApproveTime(insertTime).
				Language(migrate.GetJudgeLanguageByCodeOJ(language)).
				Code(code.String).
				CodeLength(length).
				Status(migrate.GetJudgeStatusByCodeOJ(result)).
				Score(score).
				JudgeTime(judgeTime.Time).
				Judger(judger).
				Build()

			judgeJobs = append(judgeJobs, judgeJob)

			mongoStatusId++
		}

		if !hasData {
			break // 这一批没有数据了，结束循环
		}

		err = foundationdao.GetJudgeJobDao().UpdateJudgeJobs(ctx, judgeJobs)
		if err != nil {
			return err
		}
		judgeJobs = nil

		slog.Info("migrate judge job", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return nil
}
