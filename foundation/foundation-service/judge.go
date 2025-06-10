package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	metatime "meta/meta-time"
	"meta/singleton"
	"slices"
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

func (s *JudgeService) CheckJudgeViewAuth(ctx *gin.Context, id int) (
	int,
	bool,
	*foundationmodel.ContestViewLock,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageJudge)
	if err != nil {
		return userId, false, nil, err
	}
	judgeAuth, err := foundationdao.GetJudgeJobDao().GetJudgeJobViewAuth(ctx, id)
	if err != nil {
		return userId, false, nil, err
	}
	if judgeAuth == nil {
		return userId, false, nil, nil
	}
	if !hasAuth {
		if judgeAuth.Private {
			if judgeAuth.AuthorId != userId {
				return userId, false, nil, nil
			}
		}
	}
	// 如果在比赛中，则以比赛中的权限为准进行一次拦截，即使具有管理源码的权限也无效
	var contest *foundationmodel.ContestViewLock
	if judgeAuth.ContestId > 0 {
		contest, err = foundationdao.GetContestDao().GetContestViewLock(ctx, judgeAuth.ContestId)
		if err != nil {
			return userId, false, contest, err
		}
		nowTime := metatime.GetTimeNow()
		hasStatusAuth, hasDetailAuth := s.isContestJudgeHasViewAuth(
			contest, userId,
			nowTime,
			judgeAuth.AuthorId,
			&judgeAuth.ApproveTime,
		)
		if !hasStatusAuth || !hasDetailAuth {
			return userId, false, contest, nil
		}
	}
	return userId, true, contest, nil
}

func (s *JudgeService) GetJudge(ctx context.Context, id int) (*foundationmodel.JudgeJob, error) {
	judgeJob, err := foundationdao.GetJudgeJobDao().GetJudgeJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if judgeJob == nil {
		return nil, nil
	}
	judgerName, err := foundationdao.GetJudgerDao().GetJudgerName(ctx, judgeJob.Judger)
	if err != nil {
		return nil, err
	}
	judgeJob.JudgerName = judgerName
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

func (s *JudgeService) GetJudgeCode(ctx context.Context, id int) (
	foundationjudge.JudgeLanguage,
	*string,
	error,
) {
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
			nowTime := metatime.GetTimeNow()
			problemMap := make(map[string]int)
			for _, judgeJob := range judgeJobs {
				if judgeJob.ProblemId != "" {
					if index, ok := problemMap[judgeJob.ProblemId]; ok {
						judgeJob.ContestProblemIndex = index
						continue
					}
					index, err := foundationdao.GetContestDao().GetProblemIndex(ctx, contestId, judgeJob.ProblemId)
					if err != nil {
						return nil, 0, err
					}
					judgeJob.ContestProblemIndex = index
					problemMap[judgeJob.ProblemId] = index

					judgeJob.ProblemId = ""
				}
			}

			// 隐藏部分信息
			for _, judgeJob := range judgeJobs {
				if contest.Type == foundationmodel.ContestTypeAcm {
					// IOI模式之外隐藏分数信息
					if judgeJob.Score < 100 {
						judgeJob.Score = 0
					}
				}

				hasStatusAuth, hasDetailAuth := s.isContestJudgeHasViewAuth(
					contest, userId,
					nowTime,
					judgeJob.AuthorId,
					&judgeJob.ApproveTime,
				)

				if !hasStatusAuth {
					judgeJob.Status = foundationjudge.JudgeStatusUnknown
				}
				if !hasDetailAuth {
					judgeJob.Language = foundationjudge.JudgeLanguageUnknown
					judgeJob.CodeLength = 0
				}
				if !hasStatusAuth || !hasDetailAuth {
					judgeJob.Memory = 0
					judgeJob.Time = 0
				}
			}
		}
	}
	return judgeJobs, totalCount, nil
}

func (s *JudgeService) isContestJudgeHasViewAuth(
	contest *foundationmodel.ContestViewLock,
	userId int,
	nowTime time.Time,
	authorId int,
	approveTime *time.Time,
) (
	hasStatusAuth bool,
	hasDetailAuth bool,
) {
	hasStatusAuth = true
	hasDetailAuth = true

	isEnd := nowTime.After(contest.EndTime)
	hasLockDuration := contest.LockRankDuration != nil && *contest.LockRankDuration > 0
	isLocked := hasLockDuration &&
		(contest.AlwaysLock || !isEnd) &&
		approveTime.After(contest.EndTime.Add(-*contest.LockRankDuration))

	// 不需要对管理员隐藏信息
	if contest.OwnerId != userId && !slices.Contains(contest.AuthMembers, userId) {
		if isLocked {
			if contest.Type == foundationmodel.ContestTypeOi {
				hasStatusAuth = false
			} else {
				if authorId != userId {
					hasStatusAuth = false
				}
			}
		}
		if authorId != userId {
			if isLocked {
				hasDetailAuth = false
			} else {
				if !isEnd {
					hasDetailAuth = false
				}
			}
		}
	}

	return
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
