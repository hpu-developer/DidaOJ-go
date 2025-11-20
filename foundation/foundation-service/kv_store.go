package foundationservice

import (
	"context"
	"encoding/json"

	foundationdao "foundation/foundation-dao"
	"meta/singleton"
)

type KVStoreService struct {
}

var singletonKVStoreService = singleton.Singleton[KVStoreService]{}

func GetKVStoreService() *KVStoreService {
	return singletonKVStoreService.GetInstance(
		func() *KVStoreService {
			return &KVStoreService{}
		},
	)
}
func (s *KVStoreService) GetValue(ctx context.Context, key string) (json.RawMessage, error) {
	return foundationdao.GetKVStoreDao().GetValue(ctx, key)
}
