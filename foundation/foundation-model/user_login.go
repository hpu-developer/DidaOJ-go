package foundationmodel

import (
	"time"
)

type UserLogin struct {
	Id         int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	UserId     int       `json:"user_id" gorm:"index;column:user_id"`
	IP         string    `json:"ip" gorm:"type:inet;column:ip"` // PostgreSQL INET
	UserAgent  string    `json:"user_agent" gorm:"type:text;column:user_agent"`
	InsertTime time.Time `json:"insert_time" gorm:"column:insert_time"`
}

func (u *UserLogin) TableName() string {
	return "user_login"
}

type UserLoginBuilder struct {
	item *UserLogin
}

func NewUserLoginBuilder() *UserLoginBuilder {
	return &UserLoginBuilder{item: &UserLogin{}}
}

func (b *UserLoginBuilder) Id(id int) *UserLoginBuilder {
	b.item.Id = id
	return b
}

func (b *UserLoginBuilder) UserId(userId int) *UserLoginBuilder {
	b.item.UserId = userId
	return b
}

func (b *UserLoginBuilder) IP(ip string) *UserLoginBuilder {
	b.item.IP = ip
	return b
}

func (b *UserLoginBuilder) UserAgent(userAgent string) *UserLoginBuilder {
	b.item.UserAgent = userAgent
	return b
}

func (b *UserLoginBuilder) InsertTime(insertTime time.Time) *UserLoginBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *UserLoginBuilder) Build() *UserLogin {
	return b.item
}
