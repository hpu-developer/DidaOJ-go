package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
)

type MigrateProblemDailySqlService struct {
}

var singletonMigrateProblemDailySqlService = singleton.Singleton[MigrateProblemDailySqlService]{}

func GetMigrateProblemDailySqlService() *MigrateProblemDailySqlService {
	return singletonMigrateProblemDailySqlService.GetInstance(
		func() *MigrateProblemDailySqlService {
			return &MigrateProblemDailySqlService{}
		},
	)
}

func (s *MigrateProblemDailySqlService) Start(ctx context.Context) error {

	problemDailyList, err := foundationdaomongo.GetProblemDailyDao().GetProblemListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("problemDailyList", "problemDailyList", len(problemDailyList))

	for _, problemDaily := range problemDailyList {

		problemId, err := foundationdao.GetProblemDao().GetProblemIdByKey(problemDaily.ProblemId)
		if err != nil {
			return err
		}

		newProblemDaily := foundationmodel.NewProblemDailyBuilder().
			Key(problemDaily.Id).
			ProblemId(problemId).
			Solution(problemDaily.Solution).
			Code(problemDaily.Code).
			Inserter(problemDaily.CreatorId).
			Modifier(problemDaily.UpdaterId).
			InsertTime(problemDaily.CreateTime).
			ModifyTime(problemDaily.UpdateTime).
			Build()

		err = foundationdao.GetProblemDailyDao().InsertProblemDaily(ctx, newProblemDaily)
		if err != nil {
			return err
		}
	}

	return nil
}
