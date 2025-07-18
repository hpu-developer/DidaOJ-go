package foundationservice

import (
	foundationdao "foundation/foundation-dao"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"github.com/gin-gonic/gin"
	"meta/set"
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
	*foundationview.ProblemDaily,
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
	problemId int,
	page int,
	pageSize int,
) (
	dailyList []*foundationview.ProblemDailyList,
	totalCount int,
	tags []*foundationmodel.Tag,
	attemptStatus map[int]foundationenum.ProblemAttemptStatus,
	err error,
) {
	dailyList, totalCount, err = foundationdao.GetProblemDailyDao().GetDailyList(
		ctx,
		hasAuth,
		startDate,
		endDate,
		problemId,
		page,
		pageSize,
	)
	if err != nil {
		return
	}
	if len(dailyList) == 0 {
		return
	}
	problemIdSet := set.New[int]()
	for _, daily := range dailyList {
		problemIdSet.Add(daily.ProblemId)
	}
	problemIds := problemIdSet.ToSlice()
	var problemTagMap map[int][]int
	problemTagMap, err = foundationdao.GetProblemTagDao().GetProblemTagMap(ctx, problemIds)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	var tagIds []int
	for _, tag := range problemTagMap {
		tagIds = append(tagIds, tag...)
	}
	if len(tagIds) > 0 {
		tags, err = foundationdao.GetTagDao().GetTags(ctx, tagIds)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}
	var problemStatus map[int]foundationenum.ProblemAttemptStatus
	if userId > 0 {
		problemStatus, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx,
			userId,
			problemIds,
			-1,
			nil,
			nil,
		)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}
	for _, daily := range dailyList {
		tag, ok := problemTagMap[daily.ProblemId]
		if ok {
			daily.Tags = tag
		} else {
			daily.Tags = nil
		}
	}
	return dailyList, totalCount, tags, problemStatus, nil
}

func (s *ProblemDailyService) GetDailyRecently(ctx *gin.Context, userId int) (
	[]*foundationview.ProblemDailyList,
	map[int]foundationenum.ProblemAttemptStatus,
	error,
) {
	daily, err := foundationdao.GetProblemDailyDao().GetDailyRecently(ctx)
	if err != nil {
		return nil, nil, err
	}
	if len(daily) == 0 {
		return nil, nil, nil
	}
	var problemAttemptStatus map[int]foundationenum.ProblemAttemptStatus
	if userId > 0 {
		problemIds := make([]int, len(daily))
		for i, d := range daily {
			problemIds[i] = d.ProblemId
		}
		problemAttemptStatus, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx, userId, problemIds, -1, nil, nil,
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
