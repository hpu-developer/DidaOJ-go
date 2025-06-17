package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"meta/singleton"
	"slices"
	"sort"
	"strings"
	"time"
	"web/request"
)

type ProblemService struct {
}

var singletonProblemService = singleton.Singleton[ProblemService]{}

func GetProblemService() *ProblemService {
	return singletonProblemService.GetInstance(
		func() *ProblemService {
			return &ProblemService{}
		},
	)
}

func (s *ProblemService) CheckSubmitAuth(ctx *gin.Context, id string) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		problem, err := foundationdao.GetProblemDao().GetProblemViewAuth(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if problem == nil {
			return userId, false, nil
		}
		if problem.Private &&
			problem.CreatorId != userId &&
			!slices.Contains(problem.AuthUsers, userId) {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *ProblemService) GetProblemView(
	ctx context.Context,
	id string,
	userId int,
	hasAuth bool,
) (*foundationmodel.Problem, error) {
	problem, err := foundationdao.GetProblemDao().GetProblemView(ctx, id, userId, hasAuth)
	if err != nil {
		return nil, err
	}
	if problem == nil {
		return nil, nil
	}
	if problem.CreatorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, problem.CreatorId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil
		}
		problem.CreatorUsername = &user.Username
		problem.CreatorNickname = &user.Nickname
	}
	return problem, nil
}

func (s *ProblemService) GetProblemIdByContest(ctx *gin.Context, contestId int, problemIndex int) (*string, error) {
	return foundationdao.GetContestDao().GetProblemIdByContest(ctx, contestId, problemIndex)
}

func (s *ProblemService) GetProblemDescription(
	ctx context.Context,
	id string,
) (*string, error) {
	return foundationdao.GetProblemDao().GetProblemDescription(ctx, id)
}

func (s *ProblemService) GetProblemJudge(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	problem, err := foundationdao.GetProblemDao().GetProblemJudge(ctx, id)
	if err != nil {
		return nil, err
	}
	if problem == nil {
		return nil, nil
	}
	if problem.CreatorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, problem.CreatorId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil
		}
		problem.CreatorUsername = &user.Username
		problem.CreatorNickname = &user.Nickname
	}
	return problem, nil
}

func (s *ProblemService) HasProblem(ctx context.Context, id string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblem(ctx, id)
}

func (s *ProblemService) HasProblemTitle(ctx *gin.Context, title string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblemTitle(ctx, title)
}

func (s *ProblemService) GetProblemList(
	ctx context.Context,
	oj string, title string, tag string,
	page int, pageSize int,
) ([]*foundationmodel.Problem, int, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetProblemTagDao().SearchTags(ctx, tag)
		if err != nil {
			return nil, 0, err
		}
		if len(tags) == 0 {
			return nil, 0, nil
		}
	}
	return foundationdao.GetProblemDao().GetProblemList(
		ctx, oj, title, tags, false,
		-1, false,
		page, pageSize,
	)
}

func (s *ProblemService) GetProblemListWithUser(
	ctx context.Context, userId int, hasAuth bool,
	oj string, title string, tag string, private bool,
	page int, pageSize int,
) ([]*foundationmodel.Problem, int, map[string]foundationmodel.ProblemAttemptStatus, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetProblemTagDao().SearchTags(ctx, tag)
		if err != nil {
			return nil, 0, nil, err
		}
		if len(tags) == 0 {
			return nil, 0, nil, nil
		}
	}
	problemList, totalCount, err := foundationdao.GetProblemDao().GetProblemList(
		ctx, oj, title, tags, private,
		userId, hasAuth,
		page, pageSize,
	)
	if err != nil {
		return nil, 0, nil, err
	}
	if len(problemList) <= 0 {
		return nil, 0, nil, nil
	}
	var problemIds []string
	for _, problem := range problemList {
		problemIds = append(problemIds, problem.Id)
	}
	problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, userId, -1, nil, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	return problemList, totalCount, problemStatus, nil
}

func (s *ProblemService) GetProblemRecommend(
	ctx context.Context,
	userId int,
	hasAuth bool,
	problemId string,
) ([]*foundationmodel.Problem, error) {
	var err error
	var problemIds []string
	if problemId == "" {
		problemIds, err = foundationdao.GetJudgeJobDao().GetProblemRecommendByUser(ctx, userId, hasAuth)
	} else {
		problemIds, err = foundationdao.GetJudgeJobDao().GetProblemRecommendByProblem(ctx, userId, hasAuth, problemId)
	}
	if err != nil {
		return nil, err
	}
	if len(problemIds) == 0 {
		return nil, nil
	}
	sort.Slice(
		problemIds, func(a, b int) bool {
			lengthA := len(problemIds[a])
			lengthB := len(problemIds[b])
			if lengthA != lengthB {
				return lengthA < lengthB
			}
			return strings.Compare(problemIds[a], problemIds[b]) < 0
		},
	)
	problemList, err := foundationdao.GetProblemDao().GetProblems(ctx, problemIds)
	if err != nil {
		return nil, err
	}
	if len(problemList) == 0 {
		return nil, nil
	}
	return problemList, nil
}

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) (
	[]*foundationmodel.ProblemTag,
	int,
	error,
) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}

func (s *ProblemService) GetProblemTagByIds(ctx context.Context, ids []int) ([]*foundationmodel.ProblemTag, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, ids)
}

func (s *ProblemService) GetProblemTitles(ctx *gin.Context, userId int, hasAuth bool, problems []string) (
	[]*foundationmodel.ProblemViewTitle,
	error,
) {
	return foundationdao.GetProblemDao().GetProblemTitles(ctx, userId, hasAuth, problems)
}

func (s *ProblemService) FilterValidProblemIds(ctx *gin.Context, ids []string) ([]string, error) {
	return foundationdao.GetProblemDao().FilterValidProblemIds(ctx, ids)
}

func (s *ProblemService) PostCreate(ctx context.Context, userId int, requestData *request.ProblemEdit) (
	*string,
	error,
) {
	return foundationdao.GetProblemDao().PostCreate(ctx, userId, requestData)
}

func (s *ProblemService) PostEdit(ctx context.Context, userId int, requestData *request.ProblemEdit) (
	*time.Time,
	error,
) {
	return foundationdao.GetProblemDao().PostEdit(ctx, userId, requestData)
}

func (s *ProblemService) UpdateProblemJudgeInfo(
	ctx context.Context,
	id string,
	judgeType foundationjudge.JudgeType,
	md5 string,
) error {
	return foundationdao.GetProblemDao().UpdateProblemJudgeInfo(ctx, id, judgeType, md5)
}

func (s *ProblemService) GetProblemIdByDaily(ctx *gin.Context, dailyId string) (*string, error) {
	return foundationdao.GetProblemDailyDao().GetProblemIdByDaily(ctx, dailyId)
}

func (s *ProblemService) GetDailyList(
	ctx *gin.Context,
	userId int,
	startDate *string,
	endDate *string,
	page int,
	pageSize int,
) (
	[]*foundationmodel.ProblemDaily,
	int,
	[]*foundationmodel.ProblemTag,
	map[string]foundationmodel.ProblemAttemptStatus,
	error,
) {
	dailyList, totalCount, err := foundationdao.GetProblemDailyDao().GetDailyList(
		ctx,
		startDate,
		endDate,
		page,
		pageSize,
	)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	if len(dailyList) == 0 {
		return nil, 0, nil, nil, nil
	}
	var problemIds []string
	for _, daily := range dailyList {
		problemIds = append(problemIds, daily.ProblemId)
	}
	problemList, err := foundationdao.GetProblemDao().GetProblems(ctx, problemIds)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	var tagIds []int
	for _, problem := range problemList {
		tagIds = append(tagIds, problem.Tags...)
	}
	tags, err := foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, tagIds)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	var problemStatus map[string]foundationmodel.ProblemAttemptStatus
	if userId > 0 {
		problemStatus, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx,
			problemIds,
			userId,
			-1,
			nil,
			nil,
		)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}
	problemMap := make(map[string]*foundationmodel.Problem)
	for _, problem := range problemList {
		problemMap[problem.Id] = problem
	}
	for _, daily := range dailyList {
		problem, ok := problemMap[daily.ProblemId]
		if ok {
			daily.Title = &problem.Title
			daily.Tags = problem.Tags
			daily.Accept = problem.Accept
			daily.Attempt = problem.Attempt
		}
	}
	return dailyList, totalCount, tags, problemStatus, nil
}

func (s *ProblemService) GetDailyRecently(ctx *gin.Context, userId int) (
	[]*foundationmodel.ProblemDaily,
	map[string]foundationmodel.ProblemAttemptStatus,
	error,
) {
	daily, err := foundationdao.GetProblemDailyDao().GetDailyRecently(ctx)
	if err != nil {
		return nil, nil, err
	}
	if daily == nil {
		return nil, nil, nil
	}
	for _, d := range daily {
		title, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, &d.ProblemId)
		if err == nil {
			d.Title = title
		} else {
			titlePtr := "未知题目"
			d.Title = &titlePtr
		}
	}
	var problemAttemptStatus map[string]foundationmodel.ProblemAttemptStatus
	if userId > 0 {
		problemIds := make([]string, len(daily))
		for i, d := range daily {
			problemIds[i] = d.ProblemId
		}
		problemAttemptStatus, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx, problemIds, userId, -1, nil, nil,
		)
		if err != nil {
			return nil, nil, err
		}
	}
	return daily, problemAttemptStatus, nil
}
