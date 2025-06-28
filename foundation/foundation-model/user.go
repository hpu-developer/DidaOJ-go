package foundationmodel

import (
	foundationenum "foundation/foundation-enum"
	"time"
)

type User struct {
	Id           int                       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string                    `json:"username" gorm:"type:varchar(50);unique;not null;comment:用户名"`
	Nickname     string                    `json:"nickname" gorm:"type:varchar(80);not null;comment:昵称"`
	RealName     *string                   `json:"real_name,omitempty" gorm:"type:varchar(20);comment:真实名称"`
	Password     string                    `json:"password" gorm:"type:char(80);not null"`
	Email        string                    `json:"email,omitempty" gorm:"type:varchar(90)"`
	Gender       foundationenum.UserGender `json:"gender,omitempty" gorm:"type:tinyint(1);comment:性别"`
	Number       *string                   `json:"number,omitempty" gorm:"type:varchar(20);comment:身份标识"`
	Slogan       *string                   `json:"slogan,omitempty" gorm:"type:varchar(50);comment:签名"`
	Organization *string                   `json:"organization,omitempty" gorm:"type:varchar(80);comment:组织"`
	QQ           *string                   `json:"qq,omitempty" gorm:"type:varchar(15);comment:QQ"`
	VjudgeId     *string                   `json:"vjudge_id,omitempty" gorm:"type:varchar(15);comment:VjudgeId"`
	Github       *string                   `json:"github,omitempty" gorm:"type:varchar(15);comment:Github"`
	Codeforces   *string                   `json:"codeforces,omitempty" gorm:"type:varchar(20)"`
	CheckInCount int                       `json:"check_in_count,omitempty" gorm:"comment:签到次数"`
	InsertTime   time.Time                 `json:"insert_time" gorm:"type:datetime;not null"`
	ModifyTime   time.Time                 `json:"modify_time" gorm:"type:datetime;not null"`
	Accept       int                       `json:"accept,omitempty"`
	Attempt      int                       `json:"attempt,omitempty"`
}

func (u *User) TableName() string {
	return "user"
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

func (b *UserBuilder) RealName(realName *string) *UserBuilder {
	b.item.RealName = realName
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

func (b *UserBuilder) Gender(gender foundationenum.UserGender) *UserBuilder {
	b.item.Gender = gender
	return b
}

func (b *UserBuilder) Number(number *string) *UserBuilder {
	b.item.Number = number
	return b
}

func (b *UserBuilder) Slogan(slogan *string) *UserBuilder {
	b.item.Slogan = slogan
	return b
}

func (b *UserBuilder) Organization(organization *string) *UserBuilder {
	b.item.Organization = organization
	return b
}

func (b *UserBuilder) QQ(qq *string) *UserBuilder {
	b.item.QQ = qq
	return b
}

func (b *UserBuilder) VjudgeId(vjudgeId *string) *UserBuilder {
	b.item.VjudgeId = vjudgeId
	return b
}

func (b *UserBuilder) Github(github *string) *UserBuilder {
	b.item.Github = github
	return b
}

func (b *UserBuilder) Codeforces(codeforces *string) *UserBuilder {
	b.item.Codeforces = codeforces
	return b
}

func (b *UserBuilder) CheckInCount(checkInCount int) *UserBuilder {
	b.item.CheckInCount = checkInCount
	return b
}

func (b *UserBuilder) InsertTime(insertTime time.Time) *UserBuilder {
	b.item.InsertTime = insertTime
	return b
}

func (b *UserBuilder) ModifyTime(modifyTime time.Time) *UserBuilder {
	b.item.ModifyTime = modifyTime
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

func (b *UserBuilder) Build() *User {
	return b.item
}
