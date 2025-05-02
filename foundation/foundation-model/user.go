package foundationmodel

import (
	"time"
)

type User struct {
	Id           int       `json:"id" bson:"_id"`                                        // 数据库索引时真正的Id
	Username     string    `json:"username" bson:"username"`                             // 对用户展示的唯一标识
	Nickname     string    `json:"nickname,omitempty" bson:"nickname"`                   // 显示的昵称
	Password     string    `json:"password" bson:"password"`                             // 密码
	Email        string    `json:"email" bson:"email"`                                   // 邮箱
	Sign         string    `json:"sign,omitempty" bson:"sign,omitempty"`                 // 签名
	Organization string    `json:"organization,omitempty" bson:"organization,omitempty"` // 组织
	RegTime      time.Time `json:"reg_time" bson:"reg_time"`                             // 注册时间
}

type UserAccountInfo struct {
	Id       int    `json:"id" bson:"_id"`                      // 数据库索引时真正的Id
	Username string `json:"username" bson:"username"`           // 对用户展示的唯一标识
	Nickname string `json:"nickname,omitempty" bson:"nickname"` // 显示的昵称
}

type UserLogin struct {
	Id       int    `json:"id" bson:"_id"`                      // 数据库索引时真正的Id
	Username string `json:"username" bson:"username"`           // 对用户展示的唯一标识
	Nickname string `json:"nickname,omitempty" bson:"nickname"` // 显示的昵称
	Password string `json:"password" bson:"password"`           // 密码
}

type UserBuilder struct {
	item *User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{item: &User{}}
}

func (b *UserBuilder) Id(id int) *UserBuilder {
	b.item.Id = id
	return b
}

func (b *UserBuilder) Username(username string) *UserBuilder {
	b.item.Username = username
	return b
}

func (b *UserBuilder) Nickname(nickname string) *UserBuilder {
	b.item.Nickname = nickname
	return b
}

func (b *UserBuilder) Password(password string) *UserBuilder {
	b.item.Password = password
	return b
}

func (b *UserBuilder) Email(email string) *UserBuilder {
	b.item.Email = email
	return b
}

func (b *UserBuilder) Sign(sign string) *UserBuilder {
	b.item.Sign = sign
	return b
}

func (b *UserBuilder) Organization(organization string) *UserBuilder {
	b.item.Organization = organization
	return b
}

func (b *UserBuilder) RegTime(regTime time.Time) *UserBuilder {
	b.item.RegTime = regTime
	return b
}

func (b *UserBuilder) Build() *User {
	return b.item
}
