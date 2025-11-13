package foundationmodel

import (
	foundationuser "foundation/foundation-user"
	"time"
)

// UserCoin 用户金币记录模型
type UserCoin struct {
	UserId       int                       `json:"user_id" gorm:"column:user_id;not null;index:idx_user_type_param,unique"`
	Value        int                       `json:"value" gorm:"column:value;not null;comment:金币变化量（正值增加，负值减少）"`
	InserterTime time.Time                 `json:"inserter_time" gorm:"column:inserter_time;not null;default:CURRENT_TIMESTAMP"`
	Type         foundationuser.CoinType   `json:"type" gorm:"column:type;type:int;not null;index:idx_user_type_param,unique;comment:金币类型（如：check_in, reward等）"`
	Param        string                    `json:"param" gorm:"column:param;type:varchar(255);not null;index:idx_user_type_param,unique;comment:参数（如：活动ID等）"`
}

// TableName 指定表名
func (u *UserCoin) TableName() string {
	return "user_coin"
}

// UserCoinBuilder 构建器模式
type UserCoinBuilder struct {
	item *UserCoin
}

// NewUserCoinBuilder 创建新的构建器
func NewUserCoinBuilder() *UserCoinBuilder {
	return &UserCoinBuilder{
		item: &UserCoin{},
	}
}

// UserId 设置用户ID
func (b *UserCoinBuilder) UserId(userId int) *UserCoinBuilder {
	b.item.UserId = userId
	return b
}

// Value 设置金币变化量
func (b *UserCoinBuilder) Value(value int) *UserCoinBuilder {
	b.item.Value = value
	return b
}

// InserterTime 设置插入时间
func (b *UserCoinBuilder) InserterTime(insertTime time.Time) *UserCoinBuilder {
	b.item.InserterTime = insertTime
	return b
}

// Type 设置金币类型
func (b *UserCoinBuilder) Type(coinType foundationuser.CoinType) *UserCoinBuilder {
	b.item.Type = coinType
	return b
}

// Param 设置参数
func (b *UserCoinBuilder) Param(param string) *UserCoinBuilder {
	b.item.Param = param
	return b
}

// Build 构建模型实例
func (b *UserCoinBuilder) Build() *UserCoin {
	return b.item
}