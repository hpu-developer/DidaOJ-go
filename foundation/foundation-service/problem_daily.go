package foundationservice

import (
	foundationdao "foundation/foundation-dao"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"github.com/gin-gonic/gin"
	"meta/singleton"
)

type ProblemDailyService struct {
}

var singletonProblemDailyService = singleton.Singleton[ProblemDailyService]{}

func GetProblemDailyService() *ProblemDailyService {
	return singletonProblemDailyService.GetInstance(
		func() *ProblemDailyService {
			return &ProblemDailyService{}
		},
	)
}

func (s *ProblemDailyService) HasProblemDaily(ctx *gin.Context, dailyId string) (bool, error) {
	return foundationdao.GetProblemDailyDao().HasProblemDaily(ctx, dailyId)
}

func (s *ProblemDailyService) HasProblemDailyProblem(ctx *gin.Context, problemId int) (bool, error) {
	return foundationdao.GetProblemDailyDao().HasProblemDailyProblem(ctx, problemId)
}

func (s *ProblemDailyService) GetProblemDaily(ctx *gin.Context, dailyId string, hasAuth bool) (
	*foundationmodel.ProblemDaily,
	error,
) {
	return foundationdao.GetProblemDailyDao().GetProblemDaily(ctx, dailyId, hasAuth)
}

func (s *ProblemDailyService) GetProblemDailyEdit(ctx *gin.Context, dailyId string) (
	*foundationview.ProblemDailyEdit,
	error,
) {
	return foundationdao.GetProblemDailyDao().GetProblemDailyEdit(ctx, dailyId)
}

func (s *ProblemDailyService) GetDailyList(
	ctx *gin.Context,
	userId int,
	hasAuth bool,
	startDate *string,
	endDate *string,
	problemId string,
	page int,
	pageSize int,
) (
	[]*foundationmodel.ProblemDaily,
	int,
	[]*foundationmodel.ProblemTag,
	map[string]foundationenum.ProblemAttemptStatus,
	error,
) {
	dailyList, totalCount, err := foundationdao.GetProblemDailyDao().GetDailyList(
		ctx,
		hasAuth,
		startDate,
		endDate,
		problemId,
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
	var tags []*foundationmodel.ProblemTag
	if len(tagIds) > 0 {
		tags, err = foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, tagIds)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}
	var problemStatus map[string]foundationenum.ProblemAttemptStatus
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

func (s *ProblemDailyService) GetDailyRecently(ctx *gin.Context, userId int) (
	[]*foundationmodel.ProblemDaily,
	map[string]foundationenum.ProblemAttemptStatus,
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
	var problemAttemptStatus map[string]foundationenum.ProblemAttemptStatus
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

func (s *ProblemDailyService) PostDailyCreate(
	ctx *gin.Context,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	return foundationdao.GetProblemDailyDao().InsertProblemDaily(ctx, problemDaily)
}

func (s *ProblemDailyService) PostDailyEdit(
	ctx *gin.Context,
	id string,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	return foundationdao.GetProblemDailyDao().UpdateProblemDaily(ctx, id, problemDaily)
}
