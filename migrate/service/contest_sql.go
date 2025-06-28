package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
	"time"
)

type MigrateContestSqlService struct {
}

var singletonMigrateContestSqlService = singleton.Singleton[MigrateContestSqlService]{}

func GetMigrateContestSqlService() *MigrateContestSqlService {
	return singletonMigrateContestSqlService.GetInstance(
		func() *MigrateContestSqlService {
			return &MigrateContestSqlService{}
		},
	)
}

func (s *MigrateContestSqlService) Start(ctx context.Context) error {

	contestList, err := foundationdaomongo.GetContestDao().GetListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("contestList", "contestList", len(contestList))

	for _, problemDaily := range contestList {

		newContest := foundationmodel.NewContestBuilder().
			Id(problemDaily.Id).
			Title(problemDaily.Title).
			StartTime(problemDaily.StartTime).
			EndTime(problemDaily.EndTime).
			Inserter(problemDaily.OwnerId).
			Modifier(problemDaily.OwnerId).
			InsertTime(problemDaily.CreateTime).
			ModifyTime(problemDaily.UpdateTime).
			Private(problemDaily.Private).
			Password(problemDaily.Password).
			SubmitAnytime(problemDaily.SubmitAnytime).
			Type(problemDaily.Type).
			ScoreType(problemDaily.ScoreType).
			LockRankDuration(problemDaily.LockRankDuration).
			AlwaysLock(problemDaily.AlwaysLock).
			DiscussType(problemDaily.DiscussType).
			Build()

		if problemDaily.Description != "" {
			newContest.Description = &problemDaily.Description
		}

		if problemDaily.Notification != "" {
			newContest.Notification = &problemDaily.Notification
		}
		if newContest.ModifyTime == (time.Time{}) {
			newContest.ModifyTime = newContest.InsertTime
		}

		var contestProblems []*foundationmodel.ContestProblem
		for _, problem := range problemDaily.Problems {
			problemId, err := foundationdao.GetProblemDao().GetProblemIdByKey(problem.ProblemId)
			if err != nil {
				return err
			}
			contestProblems = append(
				contestProblems, foundationmodel.NewContestProblemBuilder().
					ProblemId(problemId).
					Index(uint8(problem.Index)).
					Score(problem.Weight).
					Build(),
			)
		}
		err = foundationdao.GetContestDao().InsertContest(
			ctx, newContest, contestProblems,
			nil,
			problemDaily.Authors,
			problemDaily.Members,
			problemDaily.AuthMembers,
			problemDaily.VMembers,
			problemDaily.Volunteers,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
