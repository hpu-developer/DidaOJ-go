package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"meta/singleton"
	"time"
)

type JudgeService struct {
}

var singletonJudgeService = singleton.Singleton[JudgeService]{}

func GetJudgeService() *JudgeService {
	return singletonJudgeService.GetInstance(
		func() *JudgeService {
			return &JudgeService{}
		},
	)
}

func (s *JudgeService) GetJudge(ctx context.Context, id int) (*foundationmodel.JudgeJob, error) {
	judgeJob, err := foundationdao.GetJudgeJobDao().GetJudgeJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if judgeJob == nil {
		return nil, nil
	}
	user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, judgeJob.AuthorId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	judgeJob.AuthorUsername = &user.Username
	judgeJob.AuthorNickname = &user.Nickname
	return judgeJob, nil
}

func (s *JudgeService) GetJudgeCode(ctx context.Context, id int) (foundationjudge.JudgeLanguage, *string, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeCode(ctx, id)
}

func (s *JudgeService) GetJudgeList(
	ctx context.Context, userId int,
	contestId int, contestProblemIndex int,
	problemId string,
	username string, language foundationjudge.JudgeLanguage, status foundationjudge.JudgeStatus, page int, pageSize int,
) ([]*foundationmodel.JudgeJob, int, error) {
	var err error
	searchUserId := -1
	if username != "" {
		searchUserId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, username)
		if err != nil {
			return nil, 0, err
		}
		if searchUserId <= 0 {
			return nil, 0, nil
		}
	}
	if contestId > 0 {
		// 计算ProblemId
		if contestProblemIndex > 0 {
			problemIdPtr, err := foundationdao.GetContestDao().GetProblemIdByContest(
				ctx,
				contestId,
				contestProblemIndex,
			)
			if err != nil {
				return nil, 0, err
			}
			if problemIdPtr == nil {
				return nil, 0, nil
			}
			problemId = *problemIdPtr
		}
	}

	judgeJobs, totalCount, err := foundationdao.GetJudgeJobDao().GetJudgeJobList(
		ctx,
		contestId,
		problemId,
		searchUserId,
		language,
		status,
		page,
		pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	if len(judgeJobs) > 0 {
		var userIds []int
		for _, judgeJob := range judgeJobs {
			userIds = append(userIds, judgeJob.AuthorId)
		}
		users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, judgeJob := range judgeJobs {
			if user, ok := userMap[judgeJob.AuthorId]; ok {
				judgeJob.AuthorUsername = &user.Username
				judgeJob.AuthorNickname = &user.Nickname
			}
		}

		if contestId > 0 {
			contest, err := foundationdao.GetContestDao().GetContestViewLock(ctx, contestId)
			if err != nil {
				return nil, 0, err
			}
			if contest == nil {
				return nil, 0, nil
			}
			if len(contest.Problems) > 0 {
				problemMap := make(map[string]int)
				for _, problem := range contest.Problems {
					if problem.ProblemId != "" {
						problemMap[problem.ProblemId] = problem.Index
					}
				}
				for _, judgeJob := range judgeJobs {
					if judgeJob.ProblemId != "" {
						judgeJob.ContestProblemIndex = problemMap[judgeJob.ProblemId]
					}
				}
			}
			for _, judgeJob := range judgeJobs {
				// 隐藏真实的ProblemId
				judgeJob.ProblemId = ""
			}

			//计算锁榜来隐藏结果信息
			//contest.StartTime
			//contest.EndTime
		}
	}
	return judgeJobs, totalCount, nil
}

func (s *JudgeService) GetRankAcProblem(
	ctx *gin.Context,
	approveStartTime *time.Time,
	approveEndTime *time.Time,
	page int,
	pageSize int,
) ([]*foundationmodel.UserRank, int, error) {
	rankUsers, totalCount, err := foundationdao.GetJudgeJobDao().GetRankAcProblem(
		ctx,
		approveStartTime,
		approveEndTime,
		page,
		pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	if len(rankUsers) > 0 {
		var userIds []int
		for _, rankUser := range rankUsers {
			userIds = append(userIds, rankUser.Id)
		}
		users, err := foundationdao.GetUserDao().GetUsersRankInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserRankInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, rankUser := range rankUsers {
			if user, ok := userMap[rankUser.Id]; ok {
				rankUser.Username = user.Username
				rankUser.Nickname = user.Nickname
				rankUser.Slogan = user.Slogan
			}
		}
	}
	return rankUsers, totalCount, nil
}

func (s *JudgeService) GetUserAcProblemIds(ctx context.Context, userId int) ([]string, error) {
	problemIds, err := foundationdao.GetJudgeJobDao().GetUserAcProblemIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	return problemIds, nil
}

func (s *JudgeService) UpdateJudge(ctx context.Context, id int, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().UpdateJudgeJob(ctx, id, judgeJob)
}

func (s *JudgeService) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
}

func (s *JudgeService) RejudgeJob(ctx context.Context, id int) error {
	return foundationdao.GetJudgeJobDao().RejudgeJob(ctx, id)
}

func (s *JudgeService) PostRejudgeSearch(
	ctx context.Context,
	id string,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
) error {
	return foundationdao.GetJudgeJobDao().RejudgeSearch(ctx, id, language, status)
}

func (s *JudgeService) RejudgeRecently(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeRecently(ctx)
}

func (s *JudgeService) RejudgeAll(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeAll(ctx)
}
