package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	metatime "meta/meta-time"
	"meta/singleton"
	"time"
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

func (s *ContestService) CheckEditAuth(ctx *gin.Context, id int) (
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

func (s *ContestService) CheckViewAuth(ctx *gin.Context, id int) (int, bool, error) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetContestDao().HasContestViewAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		return userId, hasAuth, nil
	}
	return userId, true, nil
}

func (s *ContestService) CheckSubmitAuth(ctx *gin.Context, id int) (int, bool, error) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetContestDao().HasContestSubmitAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		return userId, hasAuth, nil
	}
	return userId, true, nil
}

func (s *ContestService) HasContestTitle(ctx *gin.Context, userId int, title string) (bool, error) {
	return foundationdao.GetContestDao().HasContestTitle(ctx, userId, title)
}

func (s *ContestService) GetContest(ctx *gin.Context, id int, nowTime time.Time) (
	*foundationmodel.Contest,
	bool, bool,
	map[int]foundationmodel.ProblemAttemptStatus,
	error,
) {
	contest, err := foundationdao.GetContestDao().GetContest(ctx, id)
	if err != nil {
		return nil, false, false, nil, err
	}
	if contest == nil {
		return nil, false, false, nil, nil
	}
	ownerUser, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, contest.OwnerId)
	if err != nil {
		return nil, false, false, nil, err
	}
	contest.OwnerUsername = &ownerUser.Username
	contest.OwnerNickname = &ownerUser.Nickname

	hasAuth := true
	needPassword := contest.Password != nil
	contest.Password = nil

	var userId int

	if nowTime.Before(contest.StartTime) {
		contest.Problems = nil
	} else {
		userId, hasAuth, err = GetContestService().CheckViewAuth(ctx, id)
		if !hasAuth {
			contest.Problems = nil
		}
	}

	var attemptStatusesMap map[int]foundationmodel.ProblemAttemptStatus

	if len(contest.Problems) > 0 {
		contestProblems := map[string]*foundationmodel.ContestProblem{}
		for _, problem := range contest.Problems {
			contestProblems[problem.ProblemId] = problem
		}
		var problemIds []string
		for _, problem := range contest.Problems {
			problemIds = append(problemIds, problem.ProblemId)
		}
		problems, err := foundationdao.GetProblemDao().GetProblemListTitle(ctx, problemIds)
		if err != nil {
			return nil, false, false, nil, err
		}
		for _, problem := range problems {
			if contestProblem, ok := contestProblems[problem.Id]; ok {
				contestProblem.Title = &problem.Title
			}
		}
		judgeAccepts, err := foundationdao.GetJudgeJobDao().GetProblemContestViewAttempt(ctx, id, problemIds)
		if err != nil {
			return nil, false, false, nil, err
		}
		for _, judgeAccept := range judgeAccepts {
			if contestProblem, ok := contestProblems[judgeAccept.Id]; ok {
				contestProblem.Accept = judgeAccept.Accept
				contestProblem.Attempt = judgeAccept.Attempt
			}
		}
		if userId > 0 {
			problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, userId, id)
			if err != nil {
				return nil, false, false, nil, err
			}
			attemptStatusesMap = make(map[int]foundationmodel.ProblemAttemptStatus)
			problemIdMap := make(map[string]int)
			for _, problem := range contest.Problems {
				problemIdMap[problem.ProblemId] = problem.Index
			}
			for problemId, attemptStatus := range problemStatus {
				if index, ok := problemIdMap[problemId]; ok {
					attemptStatusesMap[index] = attemptStatus
				}
			}
		}

		// 隐藏真实题目Id
		for _, problem := range contest.Problems {
			problem.ProblemId = ""
		}
	}
	return contest, hasAuth, needPassword, attemptStatusesMap, err
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

func (s *ContestService) GetContestStartTime(ctx *gin.Context, id int) (*time.Time, error) {
	return foundationdao.GetContestDao().GetContestStartTime(ctx, id)
}

func (s *ContestService) GetContestProblems(ctx *gin.Context, id int) (
	[]int,
	error,
) {
	problems, err := foundationdao.GetContestDao().GetProblems(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(problems) == 0 {
		return nil, nil
	}
	var problemIndexes []int
	for _, problem := range problems {
		problemIndexes = append(problemIndexes, problem.Index)
	}
	return problemIndexes, nil
}

func (s *ContestService) GetContestProblemsWithAttemptStatus(ctx *gin.Context, id int, userId int) (
	[]int,
	map[int]foundationmodel.ProblemAttemptStatus,
	error,
) {
	problems, err := foundationdao.GetContestDao().GetProblems(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if len(problems) == 0 {
		return nil, nil, nil
	}
	var attemptStatusesMap map[int]foundationmodel.ProblemAttemptStatus
	if userId > 0 {
		var problemIds []string
		for _, problem := range problems {
			problemIds = append(problemIds, problem.ProblemId)
		}
		attemptStatuses, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, userId, id)
		if err != nil {
			return nil, nil, err
		}
		problemIdMap := make(map[string]int)
		for _, problem := range problems {
			problemIdMap[problem.ProblemId] = problem.Index
		}
		for problemId, attemptStatus := range attemptStatuses {
			if index, ok := problemIdMap[problemId]; ok {
				if attemptStatusesMap == nil {
					attemptStatusesMap = make(map[int]foundationmodel.ProblemAttemptStatus)
				}
				attemptStatusesMap[index] = attemptStatus
			}
		}
	}
	var problemIndexes []int
	for _, problem := range problems {
		problemIndexes = append(problemIndexes, problem.Index)
	}
	return problemIndexes, attemptStatusesMap, nil
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

func (s *ContestService) GetProblemIdByContest(ctx *gin.Context, id int, problemIndex int) (*string, error) {
	return foundationdao.GetContestDao().GetProblemIdByContest(ctx, id, problemIndex)
}

func (s *ContestService) InsertContest(ctx context.Context, contest *foundationmodel.Contest) error {
	return foundationdao.GetContestDao().InsertContest(ctx, contest)
}

func (s *ContestService) GetContestRanks(ctx context.Context, id int) (
	*foundationmodel.ContestViewRank,
	[]int,
	[]*foundationmodel.ContestRank,
	error,
) {
	contest, err := foundationdao.GetContestDao().GetContestViewRank(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	problemMap := make(map[string]int)
	for _, problem := range contest.Problems {
		problemMap[problem.ProblemId] = problem.Index
	}
	nowTime := metatime.GetTimeNow()

	isEnd := nowTime.After(contest.EndTime)
	hasLockDuration := contest.LockRankDuration != nil && *contest.LockRankDuration > 0
	isLocked := hasLockDuration &&
		(contest.AlwaysLock || !isEnd) &&
		nowTime.After(contest.EndTime.Add(-*contest.LockRankDuration))

	contestRanks, err := foundationdao.GetJudgeJobDao().GetContestRanks(
		ctx, id,
		contest.StartTime,
		lockTimePtr,
		problemMap,
	)
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
	for _, problem := range contest.Problems {
		problemIndexes = append(problemIndexes, problem.Index)
	}
	// 隐藏题目详细信息
	contest.Problems = nil
	return contest, problemIndexes, contestRanks, nil
}

func (s *ContestService) UpdateContest(ctx *gin.Context, id int, contest *foundationmodel.Contest) error {
	return foundationdao.GetContestDao().UpdateContest(ctx, id, contest)
}

func (s *ContestService) PostPassword(ctx *gin.Context, userId int, contestId int, password string) (bool, error) {
	return foundationdao.GetContestDao().PostPassword(ctx, userId, contestId, password)
}
