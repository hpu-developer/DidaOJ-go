package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
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

func (s *ContestService) CheckUserAuth(ctx *gin.Context, id int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		ownerId, err := foundationdao.GetContestDao().GetContestOwnerId(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if ownerId != userId {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *ContestService) HasContestTitle(ctx *gin.Context, userId int, title string) (bool, error) {
	return foundationdao.GetContestDao().HasContestTitle(ctx, userId, title)
}

func (s *ContestService) GetContest(ctx context.Context, id int) (*foundationmodel.Contest, error) {
	contest, err := foundationdao.GetContestDao().GetContest(ctx, id)
	if err != nil {
		return nil, err
	}
	if contest == nil {
		return nil, nil
	}
	ownerUser, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, contest.OwnerId)
	if err != nil {
		return nil, err
	}
	contest.OwnerUsername = &ownerUser.Username
	contest.OwnerNickname = &ownerUser.Nickname
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
	judgeAccepts, err := foundationdao.GetJudgeJobDao().GetProblemContestViewAttempt(ctx, id, problemIds)
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

func (s *ContestService) GetContestEdit(ctx context.Context, id int) (*foundationmodel.Contest, []string, error) {
	contest, err := foundationdao.GetContestDao().GetContestEdit(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if contest == nil {
		return nil, nil, nil
	}
	ownerUser, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, contest.OwnerId)
	if err != nil {
		return nil, nil, err
	}
	contest.OwnerUsername = &ownerUser.Username
	contest.OwnerNickname = &ownerUser.Nickname
	var problemIds []string
	for _, problem := range contest.Problems {
		problemIds = append(problemIds, problem.ProblemId)
	}
	contest.Problems = nil
	return contest, problemIds, nil
}

func (s *ContestService) GetContestList(ctx context.Context, page int, pageSize int) (
	[]*foundationmodel.Contest,
	int,
	error,
) {
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

func (s *ContestService) GetContestRanks(ctx context.Context, id int) (
	*foundationmodel.ContestRankView,
	[]int,
	[]*foundationmodel.ContestRank,
	error,
) {
	contestView, err := foundationdao.GetContestDao().GetContestRankView(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	problemMap := make(map[string]int)
	for _, problem := range contestView.Problems {
		problemMap[problem.ProblemId] = problem.Index
	}
	contestRanks, err := foundationdao.GetJudgeJobDao().GetContestRanks(ctx, id, contestView.StartTime, problemMap)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(contestRanks) > 0 {
		var userIds []int
		for _, contestRank := range contestRanks {
			userIds = append(userIds, contestRank.AuthorId)
		}
		users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, nil, nil, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, contestRank := range contestRanks {
			if user, ok := userMap[contestRank.AuthorId]; ok {
				contestRank.AuthorUsername = &user.Username
				contestRank.AuthorNickname = &user.Nickname
			}
		}
	}
	var problemIndexes []int
	for _, problem := range contestView.Problems {
		problemIndexes = append(problemIndexes, problem.Index)
	}
	// 隐藏题目详细信息
	contestView.Problems = nil
	return contestView, problemIndexes, contestRanks, nil
}

func (s *ContestService) UpdateContest(ctx *gin.Context, id int, contest *foundationmodel.Contest) error {
	return foundationdao.GetContestDao().UpdateContest(ctx, id, contest)
}
