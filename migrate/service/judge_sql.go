package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
)

type MigrateJudgeSqlService struct {
}

var singletonMigrateJudgeSqlService = singleton.Singleton[MigrateJudgeSqlService]{}

func GetMigrateJudgeSqlService() *MigrateJudgeSqlService {
	return singletonMigrateJudgeSqlService.GetInstance(
		func() *MigrateJudgeSqlService {
			return &MigrateJudgeSqlService{}
		},
	)
}

func (s *MigrateJudgeSqlService) Start(ctx context.Context) error {

	judgeJobList, err := foundationdaomongo.GetJudgeJobDao().GetListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("judgeJobList", "judgeJobList", len(judgeJobList))

	for _, judgeJob := range judgeJobList {
		if judgeJob.ProblemId == "VJUDGE-51Nod-1000" {
			continue
		}
		problemId, err := foundationdao.GetProblemDao().GetProblemIdByKey(judgeJob.ProblemId)
		if err != nil {
			return err
		}

		newJudge := foundationmodel.NewJudgeJobBuilder().
			Id(judgeJob.Id).
			ProblemId(problemId).
			Language(judgeJob.Language).
			Code(judgeJob.Code).
			CodeLength(judgeJob.CodeLength).
			Status(judgeJob.Status).
			Judger(nil).
			TaskCurrent(nil).
			TaskTotal(nil).
			Score(0).
			Time(0).
			Memory(0).
			Private(judgeJob.Private).
			RemoteJudgeId(judgeJob.RemoteJudgeId).
			RemoteAccountId(judgeJob.RemoteAccountId).
			Inserter(judgeJob.AuthorId).
			InsertTime(judgeJob.ApproveTime).
			Build()

		if judgeJob.ContestId > 0 {
			newJudge.ContestId = &judgeJob.ContestId
		} else {
			newJudge.ContestId = nil
		}
		if judgeJob.JudgeTime != nil && !judgeJob.JudgeTime.IsZero() {
			newJudge.JudgeTime = judgeJob.JudgeTime
		} else {
			newJudge.JudgeTime = nil
		}

		err = foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, newJudge)
		if err != nil {
			return err
		}
	}

	return nil
}
