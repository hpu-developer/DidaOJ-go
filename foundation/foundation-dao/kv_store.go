package foundationdao

import (
	"context"
	"encoding/json"
	"errors"
	foundationmodel "foundation/foundation-model"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (d *KVStoreDao) SetValue(ctx context.Context, key string, value json.RawMessage, expiration time.Duration) error {
	// 如果key存在则更新信息，否则插入
	// 使用GORM的Upsert功能，通过OnConflict子句实现一次数据库操作完成

	// 构建插入数据的map
	data := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	// 构建OnConflict子句
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
	}

	// 处理过期时间
	if expiration > 0 {
		// 过期时间大于0，使用数据库时间作为基准计算过期时间
		expireExpr := gorm.Expr("NOW() + ?::interval", expiration.String())
		data["expire_time"] = expireExpr
		// 更新时也使用相同的表达式
		onConflict.DoUpdates = clause.Assignments(map[string]interface{}{
			"value":       value,
			"expire_time": expireExpr,
		})
	} else {
		// 过期时间为0，设置为nil（永不过期）
		data["expire_time"] = nil
		// 更新时也设置为nil
		onConflict.DoUpdates = clause.Assignments(map[string]interface{}{
			"value":       value,
			"expire_time": nil,
		})
	}

	return d.db.WithContext(ctx).
		Model(&foundationmodel.KVStore{}).
		Clauses(onConflict).
		Create(data).Error
}

func (d *KVStoreDao) GetValue(ctx context.Context, key string) (*json.RawMessage, error) {
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
	return &kvStore.Value, nil
}

func (d *KVStoreDao) DeleteValue(ctx context.Context, key string) error {
	return d.db.WithContext(ctx).
		Where("key = ?", key).
		Delete(&foundationmodel.KVStore{}).Error
}

// SetNXValue 尝试设置一个键值对，如果键已存在但未过期，则认为失败；如果键已存在且已过期，则更新
func (d *KVStoreDao) SetNXValue(ctx context.Context, key string, value json.RawMessage, expiration time.Duration) (bool, error) {
	// 处理过期时间
	expireExpr := gorm.Expr("NULL") // 默认永不过期
	if expiration > 0 {
		// 过期时间大于0，使用数据库时间作为基准计算过期时间
		expireExpr = gorm.Expr("NOW() + ?::interval", expiration.String())
	}

	// 使用 ON CONFLICT DO UPDATE 实现原子操作，只有当记录已过期时才更新
	result := d.db.WithContext(ctx).
		Model(&foundationmodel.KVStore{}).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"value":       value,
				"expire_time": expireExpr,
			}),
			// 只有当记录已过期时才执行更新
			Where: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "kv_store.expire_time IS NOT NULL AND kv_store.expire_time <= CURRENT_TIMESTAMP"}}},
		}).
		Create(map[string]interface{}{
			"key":         key,
			"value":       value,
			"expire_time": expireExpr,
		})

	if result.Error != nil {
		return false, result.Error
	}

	// 返回是否成功插入或更新（受影响行数为1表示成功）
	return result.RowsAffected == 1, nil
}
