package service

import (
	"context"
	"meta/singleton"
)

type ProblemCrawlService struct {
}

var singletonProblemCrawlService = singleton.Singleton[ProblemCrawlService]{}

func GetProblemCrawlService() *ProblemCrawlService {
	return singletonProblemCrawlService.GetInstance(
		func() *ProblemCrawlService {
			return &ProblemCrawlService{}
		},
	)
}

func (s *ProblemCrawlService) PostCrawlProblem(ctx context.Context, oj string, id string) (*string, error) {
	switch oj {
	case "hdu":
		return GetCrawlHduService().PostCrawlProblem(ctx, id)
	case "nyoj":
		return GetCrawlNyojService().PostCrawlProblem(ctx, id)
	}
	return nil, nil
}
