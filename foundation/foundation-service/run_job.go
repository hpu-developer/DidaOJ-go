package foundationservice

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
)

type RunJobService struct {
}

var singletonRunJobService = singleton.Singleton[RunJobService]{}

func GetRunJobService() *RunJobService {
	return singletonRunJobService.GetInstance(
		func() *RunJobService {
			return &RunJobService{}
		},
	)
}

// AddRunJob 添加运行任务
func (s *RunJobService) AddRunJob(ctx context.Context, runJob *foundationmodel.RunJob) error {
	return foundationdao.GetRunJobDao().AddRunJob(ctx, runJob)
}
