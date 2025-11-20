package foundationmodel

import (
	"encoding/json"
	"time"
)

type KVStore struct {
	Key        string          `json:"key" gorm:"primaryKey;column:key"` // 主键
	Value      json.RawMessage `gorm:"column:value"`                     // 对应 JSONB
	InsertTime time.Time       `gorm:"column:insert_time"`               // 插入时间
	ExpireTime *time.Time      `gorm:"column:expire_time"`               // 可为 NULL
}

func (KVStore) TableName() string {
	return "kv_store"
}

type KVStoreBuilder struct {
	item *KVStore
}

func NewKVStoreBuilder() *KVStoreBuilder {
	return &KVStoreBuilder{
		item: &KVStore{},
	}
}

func (b *KVStoreBuilder) Key(key string) *KVStoreBuilder {
	b.item.Key = key
	return b
}

func (b *KVStoreBuilder) Value(value json.RawMessage) *KVStoreBuilder {
	b.item.Value = value
	return b
}

func (b *KVStoreBuilder) InsertTime(insertTime time.Time) *KVStoreBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *KVStoreBuilder) ExpireTime(expireTime *time.Time) *KVStoreBuilder {
	b.item.ExpireTime = expireTime
	return b
}

func (b *KVStoreBuilder) Build() *KVStore {
	return b.item
}
