package foundationservice

import (
	"context"
	"encoding/json"
	"time"

	foundationdao "foundation/foundation-dao"
	metaerror "meta/meta-error"
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

func (s *KVStoreService) GetValue(ctx context.Context, key string) (*json.RawMessage, error) {
	return foundationdao.GetKVStoreDao().GetValue(ctx, key)
}

func (s *KVStoreService) GetValueString(ctx context.Context, key string) (string, error) {
	value, err := foundationdao.GetKVStoreDao().GetValue(ctx, key)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", nil
	}
	return string(*value), nil
}

func (s *KVStoreService) SetValue(ctx context.Context, key string, value any, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return metaerror.Wrap(err, "KVStore marshal value to json failed")
	}
	return foundationdao.GetKVStoreDao().SetValue(ctx, key, jsonValue, expiration)
}
