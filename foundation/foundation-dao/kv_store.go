package foundationdao

import (
	"context"
	"encoding/json"
	"errors"
	foundationmodel "foundation/foundation-model"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
)

type KVStoreDao struct {
	db *gorm.DB
}

var singletonKVStoreDao = singleton.Singleton[KVStoreDao]{}

func GetKVStoreDao() *KVStoreDao {
	return singletonKVStoreDao.GetInstance(
		func() *KVStoreDao {
			dao := &KVStoreDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

// AddKVStore 添加运行任务
func (d *KVStoreDao) AddKVStore(ctx context.Context, KVStore *foundationmodel.KVStore) error {
	return d.db.WithContext(ctx).Create(KVStore).Error
}

// GetKVStore 获取运行任务
func (d *KVStoreDao) GetValue(ctx context.Context, key string) (json.RawMessage, error) {
	var kvStore foundationmodel.KVStore
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.KVStore{}).
		Where("key = ?", key).
		Where("expire_time IS NULL OR expire_time > CURRENT_TIMESTAMP").
		First(&kvStore).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有找到记录
		}
		return nil, err
	}
	return kvStore.Value, nil
}
