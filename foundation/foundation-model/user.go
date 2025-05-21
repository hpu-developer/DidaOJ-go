package foundationmodel

import (
	"time"
)

type User struct {
	Id           int       `json:"id" bson:"_id"`                                        // 数据库索引时真正的Id
	Username     string    `json:"username" bson:"username"`                             // 对用户展示的唯一标识
	Nickname     string    `json:"nickname,omitempty" bson:"nickname"`                   // 显示的昵称
	Password     string    `json:"password" bson:"password"`                             // 密码
	Email        string    `json:"email,omitempty" bson:"email,omitempty"`               // 邮箱
	QQ           string    `json:"qq,omitempty" bson:"qq,omitempty"`                     // QQ
	Slogan       string    `json:"slogan,omitempty" bson:"slogan,omitempty"`             // 签名
	Organization string    `json:"organization,omitempty" bson:"organization,omitempty"` // 组织
	RegTime      time.Time `json:"reg_time" bson:"reg_time"`                             // 注册时间
	Accept       int       `json:"accept" bson:"accept"`                                 // AC次数
	Attempt      int       `json:"attempt" bson:"attempt"`                               // 尝试次数
	CheckinCount int       `json:"checkin_count" bson:"checkin_count"`                   // 签到次数
	Roles        []string  `json:"roles,omitempty" bson:"roles,omitempty"`               // 角色

	// 账号关联
	VjudgeId string `json:"vjudge_id,omitempty" bson:"vjudge_id,omitempty"` // vjudge.net Id
}

type UserAccountInfo struct {
	Id       int    `json:"id" bson:"_id"`                      // 数据库索引时真正的Id
	Username string `json:"username" bson:"username"`           // 对用户展示的唯一标识
	Nickname string `json:"nickname,omitempty" bson:"nickname"` // 显示的昵称
}

type UserRankInfo struct {
	Id       int    `json:"id" bson:"_id"`                                // 数据库索引时真正的Id
	Username string `json:"username" bson:"username"`                     // 对用户展示的唯一标识
	Nickname string `json:"nickname,omitempty" bson:"nickname,omitempty"` // 显示的昵称
	Slogan   string `json:"slogan,omitempty" bson:"slogan,omitempty"`     // 显示的昵称
}

type UserLogin struct {
	Id       int      `json:"id" bson:"_id"`                          // 数据库索引时真正的Id
	Username string   `json:"username" bson:"username"`               // 对用户展示的唯一标识
	Nickname string   `json:"nickname,omitempty" bson:"nickname"`     // 显示的昵称
	Password string   `json:"password" bson:"password"`               // 密码
	Roles    []string `json:"roles,omitempty" bson:"roles,omitempty"` // 角色
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

func (b *UserBuilder) Slogan(slogan string) *UserBuilder {
	b.item.Slogan = slogan
	return b
}

func (b *UserBuilder) Organization(organization string) *UserBuilder {
	b.item.Organization = organization
	return b
}

func (b *UserBuilder) QQ(qq string) *UserBuilder {
	b.item.QQ = qq
	return b
}

func (b *UserBuilder) RegTime(regTime time.Time) *UserBuilder {
	b.item.RegTime = regTime
	return b
}

func (b *UserBuilder) Accept(accept int) *UserBuilder {
	b.item.Accept = accept
	return b
}

func (b *UserBuilder) Attempt(attempt int) *UserBuilder {
	b.item.Attempt = attempt
	return b
}

func (b *UserBuilder) VjudgeId(vjudgeId string) *UserBuilder {
	b.item.VjudgeId = vjudgeId
	return b
}

func (b *UserBuilder) CheckinCount(checkinCount int) *UserBuilder {
	b.item.CheckinCount = checkinCount
	return b
}

func (b *UserBuilder) Roles(roles []string) *UserBuilder {
	b.item.Roles = roles
	return b
}

func (b *UserBuilder) Build() *User {
	return b.item
}
