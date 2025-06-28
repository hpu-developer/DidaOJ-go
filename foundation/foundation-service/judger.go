package foundationservice

import (
	"context"
	"foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
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
	return foundationdaomongo.GetJudgerDao().GetJudgers(ctx)
}
