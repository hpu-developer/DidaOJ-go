package foundationmodel

import (
	foundationuser "foundation/foundation-user"
	"time"
)

// UserExperience 用户经验记录模型
type UserExperience struct {
	UserId       int                           `json:"user_id" gorm:"column:user_id;not null;index:idx_user_type_param,unique"`
	Value        int                           `json:"value" gorm:"column:value;not null;comment:经验值变化量（正值增加，负值减少）"`
	InserterTime time.Time                     `json:"inserter_time" gorm:"column:inserter_time;not null;default:CURRENT_TIMESTAMP"`
	Type         foundationuser.ExperienceType `json:"type" gorm:"column:type;type:int;not null;index:idx_user_type_param,unique;comment:经验类型（如：submit_ac, check_in等）"`
	Param        string                        `json:"param" gorm:"column:param;type:varchar(255);not null;index:idx_user_type_param,unique;comment:参数（如：题目ID、活动ID等）"`
}

// TableName 指定表名
func (u *UserExperience) TableName() string {
	return "user_experience"
}

// UserExperienceBuilder 构建器模式
type UserExperienceBuilder struct {
	item *UserExperience
}

// NewUserExperienceBuilder 创建新的构建器
func NewUserExperienceBuilder() *UserExperienceBuilder {
	return &UserExperienceBuilder{
		item: &UserExperience{},
	}
}

// UserId 设置用户ID
func (b *UserExperienceBuilder) UserId(userId int) *UserExperienceBuilder {
	b.item.UserId = userId
	return b
}

// Value 设置经验值变化量
func (b *UserExperienceBuilder) Value(value int) *UserExperienceBuilder {
	b.item.Value = value
	return b
}

// InserterTime 设置插入时间
func (b *UserExperienceBuilder) InserterTime(insertTime time.Time) *UserExperienceBuilder {
	b.item.InserterTime = insertTime
	return b
}

// Type 设置经验类型
func (b *UserExperienceBuilder) Type(expType foundationuser.ExperienceType) *UserExperienceBuilder {
	b.item.Type = expType
	return b
}

// Param 设置参数
func (b *UserExperienceBuilder) Param(param string) *UserExperienceBuilder {
	b.item.Param = param
	return b
}

// Build 构建模型实例
func (b *UserExperienceBuilder) Build() *UserExperience {
	return b.item
}
