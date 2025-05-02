package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
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

func (s *JudgeService) GetJudge(ctx context.Context, id string) (*foundationmodel.JudgeJob, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeJob(ctx, id)
}

func (s *JudgeService) GetJudgeList(ctx context.Context, page int, pageSize int) ([]*foundationmodel.JudgeJob, int, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeJobList(ctx, page, pageSize)
}

func (s *JudgeService) UpdateJudge(ctx context.Context, id int, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().UpdateJudgeJob(ctx, id, judgeJob)
}
