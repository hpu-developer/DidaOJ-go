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

func (s *ProblemService) GetProblemList(ctx context.Context, page int, pageSize int) ([]*foundationmodel.Problem, int, error) {
	return foundationdao.GetProblemDao().GetProblemList(ctx, page, pageSize)
}

func (s *ProblemService) GetProblemListWithUser(ctx context.Context, user string, page int, pageSize int,
) ([]*foundationmodel.Problem, int, map[string]foundationmodel.ProblemAttemptStatus, error) {
	problemList, totalCount, err := foundationdao.GetProblemDao().GetProblemList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, nil, err
	}
	var problemIds []string
	for _, problem := range problemList {
		problemIds = append(problemIds, problem.Id)
	}
	problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, user)
	if err != nil {
		return nil, 0, nil, err
	}
	return problemList, totalCount, problemStatus, nil
}

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) ([]*foundationmodel.ProblemTag, int, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}
