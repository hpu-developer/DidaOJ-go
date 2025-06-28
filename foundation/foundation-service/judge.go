package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao-mongo"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model-mongo"
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
	bool, // 是否具有查看权限
	bool, // 是否具有查看Task的权限
	*foundationmodel.ContestViewLock,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageJudge)
	if err != nil {
		return userId, false, false, nil, err
	}
	judgeAuth, err := foundationdaomongo.GetJudgeJobDao().GetJudgeJobViewAuth(ctx, id)
	if err != nil {
		return userId, false, false, nil, err
	}
	if judgeAuth == nil {
		return userId, false, false, nil, nil
	}
	if !hasAuth {
		if judgeAuth.Private {
			if judgeAuth.AuthorId != userId {
				return userId, false, false, nil, nil
			}
		}
	}
	// 如果在比赛中，则以比赛中的权限为准进行一次拦截，即使具有管理源码的权限也无效
	var contest *foundationmodel.ContestViewLock
	if judgeAuth.ContestId > 0 {
		contest, err = foundationdaomongo.GetContestDao().GetContestViewLock(ctx, judgeAuth.ContestId)
		if err != nil {
			return userId, false, false, contest, err
		}
		nowTime := metatime.GetTimeNow()
		hasStatusAuth, hasDetailAuth, hasTaskAuth := s.isContestJudgeHasViewAuth(
			contest, userId,
			nowTime,
			judgeAuth.AuthorId,
			&judgeAuth.ApproveTime,
		)
		if !hasStatusAuth || !hasDetailAuth {
			return userId, false, hasTaskAuth, contest, nil
		}
	}
	return userId, true, true, contest, nil
}

func (s *JudgeService) GetJudge(ctx context.Context, id int, fields []string) (*foundationmodel.JudgeJob, error) {
	judgeJob, err := foundationdaomongo.GetJudgeJobDao().GetJudgeJob(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	if judgeJob == nil {
		return nil, nil
	}
	judgerName, err := foundationdaomongo.GetJudgerDao().GetJudgerName(ctx, judgeJob.Judger)
	if err != nil {
		return nil, err
	}
	judgeJob.JudgerName = judgerName
	user, err := foundationdaomongo.GetUserDao().GetUserAccountInfo(ctx, judgeJob.AuthorId)
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
	return foundationdaomongo.GetJudgeJobDao().GetJudgeCode(ctx, id)
}

func (s *JudgeService) GetJudgeList(
	ctx context.Context, userId int,
	contestId int, contestProblemIndex int,
	problemId string,
	username string, language foundationjudge.JudgeLanguage, status foundationjudge.JudgeStatus, page int, pageSize int,
) ([]*foundationmodel.JudgeJob, error) {
	var err error
	searchUserId := -1
	if username != "" {
		searchUserId, err = foundationdaomongo.GetUserDao().GetUserIdByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if searchUserId <= 0 {
			return nil, nil
		}
	}
	if contestId > 0 {
		// 计算ProblemId
		if contestProblemIndex > 0 {
			problemIdPtr, err := foundationdaomongo.GetContestDao().GetProblemIdByContest(
				ctx,
				contestId,
				contestProblemIndex,
			)
			if err != nil {
				return nil, err
			}
			if problemIdPtr == nil {
				return nil, nil
			}
			problemId = *problemIdPtr
		}
	}

	judgeJobs, err := foundationdaomongo.GetJudgeJobDao().GetJudgeJobList(
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
		return nil, err
	}
	if len(judgeJobs) > 0 {
		var userIds []int
		for _, judgeJob := range judgeJobs {
			userIds = append(userIds, judgeJob.AuthorId)
		}
		users, err := foundationdaomongo.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, err
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
			contest, err := foundationdaomongo.GetContestDao().GetContestViewLock(ctx, contestId)
			if err != nil {
				return nil, err
			}
			if contest == nil {
				return nil, nil
			}
			nowTime := metatime.GetTimeNow()
			problemMap := make(map[string]int)
			for _, judgeJob := range judgeJobs {
				if judgeJob.ProblemId != "" {
					if index, ok := problemMap[judgeJob.ProblemId]; ok {
						judgeJob.ContestProblemIndex = index
						continue
					}
					index, err := foundationdaomongo.GetContestDao().GetProblemIndex(ctx, contestId, judgeJob.ProblemId)
					if err != nil {
						return nil, err
					}
					judgeJob.ContestProblemIndex = index
					problemMap[judgeJob.ProblemId] = index

					judgeJob.ProblemId = ""
				}
			}

			// 隐藏部分信息
			for _, judgeJob := range judgeJobs {
				if contest.Type == foundationenum.ContestTypeAcm {
					// IOI模式之外隐藏分数信息
					if judgeJob.Score < 100 {
						judgeJob.Score = 0
					}
				}

				hasStatusAuth, hasDetailAuth, _ := s.isContestJudgeHasViewAuth(
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
	return judgeJobs, nil
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
	hasTaskAuth bool,
) {
	// 评测状态的权限
	hasStatusAuth = true
	// 评测代码的权限
	hasDetailAuth = true
	// 评测具体任务的权限
	hasTaskAuth = true

	isEnd := nowTime.After(contest.EndTime)
	hasLockDuration := contest.LockRankDuration != nil && *contest.LockRankDuration > 0
	isLocked := hasLockDuration &&
		(contest.AlwaysLock || !isEnd) &&
		approveTime.After(contest.EndTime.Add(-*contest.LockRankDuration))

	// 不需要对管理员隐藏信息
	if contest.OwnerId != userId && !slices.Contains(contest.AuthMembers, userId) {
		if isLocked {
			if contest.Type == foundationenum.ContestTypeOi {
				hasStatusAuth = false
			} else {
				if authorId != userId {
					hasStatusAuth = false
				}
			}
			hasTaskAuth = false
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
		if !isEnd {
			hasTaskAuth = false
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
	rankUsers, totalCount, err := foundationdaomongo.GetJudgeJobDao().GetRankAcProblem(
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
		users, err := foundationdaomongo.GetUserDao().GetUsersRankInfo(ctx, userIds)
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
	problemIds, err := foundationdaomongo.GetJudgeJobDao().GetUserAcProblemIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	return problemIds, nil
}

func (s *JudgeService) GetJudgeJobCountStaticsRecently(ctx context.Context) (
	[]*foundationmodel.JudgeJobCountStatics,
	error,
) {
	return foundationdaomongo.GetJudgeJobDao().GetJudgeJobCountStaticsRecently(ctx)
}

func (s *JudgeService) GetProblemAttemptStatus(
	ctx context.Context, problemIds []string, authorId int,
	contestId int, startTime *time.Time, endTime *time.Time,
) (map[string]foundationenum.ProblemAttemptStatus, error) {
	return foundationdaomongo.GetJudgeJobDao().GetProblemAttemptStatus(
		ctx,
		problemIds,
		authorId,
		contestId,
		startTime,
		endTime,
	)
}

func (s *JudgeService) GetJudgeJobCountNotFinish(ctx context.Context) (int, error) {
	return foundationdaomongo.GetJudgeJobDao().GetJudgeJobCountNotFinish(ctx)
}

func (s *JudgeService) UpdateJudge(ctx context.Context, id int, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdaomongo.GetJudgeJobDao().UpdateJudgeJob(ctx, id, judgeJob)
}

func (s *JudgeService) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdaomongo.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
}

func (s *JudgeService) RejudgeJob(ctx context.Context, id int) error {
	return foundationdaomongo.GetJudgeJobDao().RejudgeJob(ctx, id)
}

func (s *JudgeService) PostRejudgeSearch(
	ctx context.Context,
	id string,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
) error {
	return foundationdaomongo.GetJudgeJobDao().RejudgeSearch(ctx, id, language, status)
}

func (s *JudgeService) RejudgeRecently(ctx context.Context) error {
	return foundationdaomongo.GetJudgeJobDao().RejudgeRecently(ctx)
}

func (s *JudgeService) RejudgeAll(ctx context.Context) error {
	return foundationdaomongo.GetJudgeJobDao().RejudgeAll(ctx)
}
