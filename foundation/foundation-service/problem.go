package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
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

func (s *ProblemService) GetProblem(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	return foundationdao.GetProblemDao().GetProblem(ctx, id)
}

func (s *ProblemService) GetProblemList(ctx context.Context, title string, tag string, page int, pageSize int) ([]*foundationmodel.Problem, int, error) {
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
	return foundationdao.GetProblemDao().GetProblemList(ctx, title, tags, page, pageSize)
}

func (s *ProblemService) GetProblemListWithUser(ctx context.Context, userId int, title string, tag string, page int, pageSize int,
) ([]*foundationmodel.Problem, int, map[string]foundationmodel.ProblemAttemptStatus, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetProblemTagDao().SearchTags(ctx, tag)
		if err != nil {
			return nil, 0, nil, err
		}
	}
	problemList, totalCount, err := foundationdao.GetProblemDao().GetProblemList(ctx, title, tags, page, pageSize)
	if err != nil {
		return nil, 0, nil, err
	}
	var problemIds []string
	for _, problem := range problemList {
		problemIds = append(problemIds, problem.Id)
	}
	problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, userId)
	if err != nil {
		return nil, 0, nil, err
	}
	return problemList, totalCount, problemStatus, nil
}

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) ([]*foundationmodel.ProblemTag, int, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}

func (s *ProblemService) GetProblemTagByIds(ctx context.Context, ids []int) ([]*foundationmodel.ProblemTag, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, ids)
}
