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

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) ([]*foundationmodel.ProblemTag, int, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}
