package foundationservice

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
)

type TagService struct {
}

var singletonTagService = singleton.Singleton[TagService]{}

func GetTagService() *TagService {
	return singletonTagService.GetInstance(
		func() *TagService {
			return &TagService{}
		},
	)
}

func (s *TagService) GetTagTags(ctx context.Context, ids []int) ([]*foundationmodel.Tag, error) {
	return foundationdao.GetTagDao().GetTags(ctx, ids)
}
