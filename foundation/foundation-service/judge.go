package foundationservice

import (
	"context"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationcontest "foundation/foundation-contest"
	foundationdao "foundation/foundation-dao"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationremote "foundation/foundation-remote"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metatime "meta/meta-time"
	"meta/singleton"
	"time"

	"github.com/gin-gonic/gin"
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
	*foundationview.ContestViewLock,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageJudge)
	if err != nil {
		return userId, false, false, nil, err
	}
	judgeAuth, err := foundationdao.GetJudgeJobDao().GetJudgeJobViewAuth(ctx, id)
	if err != nil {
		return userId, false, false, nil, err
	}
	if judgeAuth == nil {
		return userId, false, false, nil, nil
	}
	if !hasAuth {
		if judgeAuth.Private {
			if judgeAuth.Inserter != userId {
				return userId, false, false, nil, nil
			}
		}
	}
	// 如果在比赛中，则以比赛中的权限为准进行一次拦截，即使具有管理源码的权限也无效
	var contest *foundationview.ContestViewLock
	if judgeAuth.ContestId > 0 {
		hasAuth, err = foundationdao.GetContestDao().CheckContestEditAuth(ctx, judgeAuth.ContestId, userId)
		if err != nil {
			return userId, false, false, nil, err
		}
		if !hasAuth {
			contest, err = foundationdao.GetContestDao().GetContestViewLock(ctx, judgeAuth.ContestId)
			if err != nil {
				return userId, false, false, contest, err
			}
			nowTime := metatime.GetTimeNow()
			hasStatusAuth, hasDetailAuth, hasTaskAuth := s.isContestJudgeHasViewAuth(
				contest, userId,
				nowTime,
				judgeAuth.Inserter,
				&judgeAuth.InsertTime,
			)
			if !hasStatusAuth || !hasDetailAuth {
				return userId, false, hasTaskAuth, contest, nil
			}
		}
	}
	return userId, true, true, contest, nil
}

func (s *JudgeService) GetJudge(ctx context.Context, id int, fields []string) (*foundationview.JudgeJob, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeJob(ctx, id, fields)
}

func (s *JudgeService) GetJudgeCode(ctx context.Context, id int) (
	foundationjudge.JudgeLanguage,
	*string,
	error,
) {
	return foundationdao.GetJudgeJobDao().GetJudgeCode(ctx, id)
}

func (s *JudgeService) GetJudgeList(
	ctx context.Context,
	userId int,
	problemKey string,
	contestId int,
	username string,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
	page int,
	pageSize int,
) ([]*foundationview.JudgeJob, error) {
	var err error
	searchUserId := -1
	if username != "" {
		searchUserId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if searchUserId <= 0 {
			return nil, nil
		}
	}

	problemId := 0
	if contestId > 0 {
		if problemKey != "" {
			problemIndex := foundationcontest.GetContestProblemIndex(problemKey)
			if problemIndex <= 0 {
				return nil, metaerror.NewCode(foundationerrorcode.ParamError)
			}
			problemId, err = GetContestService().GetProblemIdByContestIndex(
				ctx,
				contestId,
				problemIndex,
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "get problem id by contest index error")
			}
		}
	} else {
		if problemKey != "" {
			problemId, err = GetProblemService().GetProblemIdByKey(ctx, problemKey)
			if err != nil {
				return nil, metaerror.Wrap(err, "get problem id by key error")
			}
		}
	}

	judgeJobs, err := foundationdao.GetJudgeJobDao().GetJudgeJobList(
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

	nowTime := metatime.GetTimeNow()

	if len(judgeJobs) > 0 {
		if contestId > 0 {
			contest, err := foundationdao.GetContestDao().GetContestViewLock(ctx, contestId)
			if err != nil {
				return nil, err
			}
			if contest == nil {
				return nil, nil
			}
			hasAuth, err := foundationdao.GetContestDao().CheckContestEditAuth(ctx, contestId, userId)
			if err != nil {
				return nil, err
			}
			if !hasAuth {
				for _, judgeJob := range judgeJobs {
					judgeJob.ProblemId = 0
					judgeJob.ProblemKey = ""
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

				if !hasAuth {
					hasStatusAuth, hasDetailAuth, _ := s.isContestJudgeHasViewAuth(
						contest, userId,
						nowTime,
						judgeJob.Inserter,
						&judgeJob.InsertTime,
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
	}
	return judgeJobs, nil
}

func (s *JudgeService) GetJudgeTaskList(ctx *gin.Context, id int) ([]*foundationmodel.JudgeTask, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeTaskList(ctx, id)
}

func (s *JudgeService) isContestJudgeHasViewAuth(
	contest *foundationview.ContestViewLock,
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

	return
}

func (s *JudgeService) IsEnableRemoteJudge(oj string, problemId string, language foundationjudge.JudgeLanguage) bool {
	if !foundationjudge.IsValidJudgeLanguage(int(language)) {
		return false
	}
	if oj == "" {
		return true
	}
	agent := foundationremote.GetRemoteAgent(foundationremote.GetRemoteTypeByString(oj))
	if agent == nil {
		return false
	}
	if !agent.IsSupportJudge(problemId, language) {
		return false
	}

	return true
}

func (s *JudgeService) GetRankAcProblem(
	ctx *gin.Context,
	approveStartTime *time.Time,
	approveEndTime *time.Time,
	page int,
	pageSize int,
) ([]*foundationview.UserRank, int, error) {
	return foundationdao.GetJudgeJobDao().GetRankAcProblem(
		ctx,
		approveStartTime,
		approveEndTime,
		page,
		pageSize,
	)
}

func (s *JudgeService) GetUserAttemptProblems(ctx context.Context, userId int) (
	[]*foundationview.ProblemViewKey,
	[]*foundationview.ProblemViewKey,
	error,
) {
	acProblems, attemptProblems, err := foundationdao.GetJudgeJobDao().GetUserAttemptProblems(ctx, userId)
	if err != nil {
		return nil, nil, err
	}
	return acProblems, attemptProblems, nil
}

func (s *JudgeService) GetJudgeJobCountStaticsRecently(ctx context.Context) (
	[]*foundationview.JudgeJobCountStatics,
	error,
) {
	return foundationdao.GetJudgeJobDao().GetJudgeJobCountStaticsRecently(ctx)
}

func (s *JudgeService) GetUserJudgeJobCountStatics(ctx context.Context, userId int, year int) (
	[]*foundationview.JudgeJobCountStatics,
	error,
) {
	return foundationdao.GetJudgeJobDao().GetUserJudgeJobCountStatics(ctx, userId, year)
}

func (s *JudgeService) GetProblemAttemptStatus(
	ctx context.Context, problemIds []int, authorId int,
	contestId int, startTime *time.Time, endTime *time.Time,
) (map[int]foundationenum.ProblemAttemptStatus, error) {
	return foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
		ctx,
		authorId,
		problemIds,
		contestId,
		startTime,
		endTime,
	)
}

func (s *JudgeService) GetProblemAttemptStatusByKey(
	ctx context.Context, problemIds []int, authorId int,
	contestId int, startTime *time.Time, endTime *time.Time,
) (map[string]foundationenum.ProblemAttemptStatus, error) {
	return foundationdao.GetJudgeJobDao().GetProblemAttemptStatusByKey(
		ctx,
		authorId,
		problemIds,
		contestId,
		startTime,
		endTime,
	)
}

func (s *JudgeService) GetJudgeJobCountNotFinish(ctx context.Context) (int, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeJobCountNotFinish(ctx)
}

func (s *JudgeService) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
}

func (s *JudgeService) RejudgeJob(ctx context.Context, id int) error {
	return foundationdao.GetJudgeJobDao().RejudgeJob(ctx, id)
}

func (s *JudgeService) PostRejudgeSearch(
	ctx context.Context,
	problemId int,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
) error {
	return foundationdao.GetJudgeJobDao().RejudgeSearch(ctx, problemId, language, status)
}

func (s *JudgeService) RejudgeRecently(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeRecently(ctx)
}

func (s *JudgeService) RejudgeAll(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeAll(ctx)
}
