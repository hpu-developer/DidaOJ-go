package foundationmodel

import (
	"strings"
	"time"
)

type UserGender int

const (
	UserGenderUnknown UserGender = 0
	UserGenderMale    UserGender = 1
	UserGenderFemale  UserGender = 2
)

type User struct {
	Id            int       `json:"id" bson:"_id"`                                            // 数据库索引时真正的Id
	Username      string    `json:"username" bson:"username"`                                 // 对用户展示的唯一标识
	UsernameLower string    `json:"username_lower,omitempty" bson:"username_lower,omitempty"` // 对用户展示的唯一标识小写，主要用于方便忽略大小写索引
	Nickname      string    `json:"nickname,omitempty" bson:"nickname"`                       // 显示的昵称
	Password      string    `json:"password" bson:"password"`                                 // 密码
	Email         string    `json:"email,omitempty" bson:"email,omitempty"`                   // 邮箱
	Number        string    `json:"number,omitempty" bson:"number,omitempty"`                 // 身份标识符
	QQ            string    `json:"qq,omitempty" bson:"qq,omitempty"`                         // QQ
	Slogan        string    `json:"slogan,omitempty" bson:"slogan,omitempty"`                 // 签名
	Organization  string    `json:"organization,omitempty" bson:"organization,omitempty"`     // 组织
	RegTime       time.Time `json:"reg_time" bson:"reg_time"`                                 // 注册时间
	Accept        int       `json:"accept" bson:"accept"`                                     // AC次数
	Attempt       int       `json:"attempt" bson:"attempt"`                                   // 尝试次数
	CheckinCount  int       `json:"checkin_count" bson:"checkin_count"`                       // 签到次数
	Roles         []string  `json:"roles,omitempty" bson:"roles,omitempty"`                   // 角色

	Gender   UserGender `json:"gender" bson:"gender"`                           // 性别，0未知，1男，2女
	RealName string     `json:"real_name,omitempty" bson:"real_name,omitempty"` // 真实姓名

	// 账号关联
	VjudgeId   string `json:"vjudge_id,omitempty" bson:"vjudge_id,omitempty"` // vjudge.net Id
	Github     string `json:"github,omitempty" bson:"github,omitempty"`       // GitHub Id
	Codeforces string `json:"codeforces,omitempty" bson:"codeforces,omitempty"`
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
	b.item.UsernameLower = strings.ToLower(username)
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

func (b *UserBuilder) Number(number string) *UserBuilder {
	b.item.Number = number
	return b
}

func (b *UserBuilder) QQ(qq string) *UserBuilder {
	b.item.QQ = qq
	return b
}

func (b *UserBuilder) RealName(realName string) *UserBuilder {
	b.item.RealName = realName
	return b
}

func (b *UserBuilder) Gender(gender UserGender) *UserBuilder {
	b.item.Gender = gender
	return b
}

func (b *UserBuilder) Github(github string) *UserBuilder {
	b.item.Github = github
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

func (b *UserBuilder) Codeforces(codeforces string) *UserBuilder {
	b.item.Codeforces = codeforces
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
