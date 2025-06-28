package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
)

type MigrateDiscussSqlService struct {
}

var singletonMigrateDiscussSqlService = singleton.Singleton[MigrateDiscussSqlService]{}

func GetMigrateDiscussSqlService() *MigrateDiscussSqlService {
	return singletonMigrateDiscussSqlService.GetInstance(
		func() *MigrateDiscussSqlService {
			return &MigrateDiscussSqlService{}
		},
	)
}

func (s *MigrateDiscussSqlService) Start(ctx context.Context) error {

	discussList, err := foundationdaomongo.GetDiscussDao().GetListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("discussList", "discussList", len(discussList))

	for _, problemDaily := range discussList {

		var problemId int
		if problemDaily.ProblemId != nil {
			problemId, err = foundationdao.GetProblemDao().GetProblemIdByKey(*problemDaily.ProblemId)
			if err != nil {
				return err
			}
		}

		newDiscuss := foundationmodel.NewDiscussBuilder().
			Id(problemDaily.Id).
			Title(problemDaily.Title).
			Content(problemDaily.Content).
			ViewCount(problemDaily.ViewCount).
			Banned(problemDaily.Banned).
			Inserter(problemDaily.AuthorId).
			InsertTime(problemDaily.InsertTime).
			Modifier(problemDaily.AuthorId).
			ModifyTime(problemDaily.ModifyTime).
			Updater(problemDaily.AuthorId).
			UpdateTime(problemDaily.UpdateTime).
			Build()

		if problemId > 0 {
			newDiscuss.ProblemId = &problemId
		} else {
			newDiscuss.ProblemId = nil
		}
		if problemDaily.ContestId > 0 {
			newDiscuss.ContestId = &problemDaily.ContestId
		} else {
			newDiscuss.ContestId = nil
		}
		if problemDaily.JudgeId > 0 {
			newDiscuss.JudgeId = &problemDaily.JudgeId
		} else {
			newDiscuss.JudgeId = nil
		}
		err = foundationdao.GetDiscussDao().InsertDiscuss(ctx, newDiscuss)
		if err != nil {
			return err
		}
	}

	return nil
}
