package foundationservice

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
)

type JudgerService struct {
}

var singletonJudgerService = singleton.Singleton[JudgerService]{}

func GetJudgerService() *JudgerService {
	return singletonJudgerService.GetInstance(
		func() *JudgerService {
			return &JudgerService{}
		},
	)
}

func (s *JudgerService) GetJudgerList(ctx context.Context) ([]*foundationmodel.Judger, error) {
	return foundationdao.GetJudgerDao().GetJudgers(ctx)
}
