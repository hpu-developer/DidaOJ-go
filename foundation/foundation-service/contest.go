package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
)

type ContestService struct {
}

var singletonContestService = singleton.Singleton[ContestService]{}

func GetContestService() *ContestService {
	return singletonContestService.GetInstance(
		func() *ContestService {
			return &ContestService{}
		},
	)
}

func (s *ContestService) GetContest(ctx context.Context, id int) (*foundationmodel.Contest, error) {
	contest, err := foundationdao.GetContestDao().GetContest(ctx, id)
	if err != nil {
		return nil, err
	}
	if contest == nil {
		return nil, nil
	}
	contestProblems := map[string]*foundationmodel.ContestProblem{}
	for _, problem := range contest.Problems {
		contestProblems[problem.ProblemId] = problem
	}
	var problemIds []string
	for _, problem := range contest.Problems {
		problemIds = append(problemIds, problem.ProblemId)
		// 隐藏题目Id
		problem.ProblemId = ""
	}
	problems, err := foundationdao.GetProblemDao().GetProblemListTitle(ctx, problemIds)
	if err != nil {
		return nil, err
	}
	for _, problem := range problems {
		if contestProblem, ok := contestProblems[problem.Id]; ok {
			contestProblem.Title = &problem.Title
		}
	}
	judgeAccepts, err := foundationdao.GetJudgeJobDao().GetProblemViewAttempt(ctx, id, problemIds)
	if err != nil {
		return nil, err
	}
	for _, judgeAccept := range judgeAccepts {
		if contestProblem, ok := contestProblems[judgeAccept.Id]; ok {
			contestProblem.Accept = judgeAccept.Accept
			contestProblem.Attempt = judgeAccept.Attempt
		}
	}
	return contest, err
}

func (s *ContestService) GetContestList(ctx context.Context, page int, pageSize int) ([]*foundationmodel.Contest, int, error) {
	contests, totalCount, err := foundationdao.GetContestDao().GetContestList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var userIds []int
	for _, contest := range contests {
		userIds = append(userIds, contest.OwnerId)
	}
	users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
	if err != nil {
		return nil, 0, err
	}
	userMap := make(map[int]*foundationmodel.UserAccountInfo)
	for _, user := range users {
		userMap[user.Id] = user
	}
	for _, contest := range contests {
		if user, ok := userMap[contest.OwnerId]; ok {
			contest.OwnerUsername = &user.Username
			contest.OwnerNickname = &user.Nickname
		}
	}
	return contests, totalCount, nil
}

func (s *ContestService) InsertContest(ctx context.Context, contest *foundationmodel.Contest) error {
	return foundationdao.GetContestDao().InsertContest(ctx, contest)
}
