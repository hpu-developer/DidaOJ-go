package foundationview

import (
	foundationenum "foundation/foundation-enum"
	"time"
)

type UserLogin struct {
	Id       int    `json:"id"`                 // 数据库索引时真正的Id
	Username string `json:"username"`           // 对用户展示的唯一标识
	Nickname string `json:"nickname,omitempty"` // 显示的昵称
	Password string `json:"password"`           // 密码

	Token *string  `json:"token,omitempty"`          // 登录令牌
	Roles []string `json:"roles,omitempty" gorm:"-"` // 角色
}

type UserInfo struct {
	Id           int                       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string                    `json:"username" gorm:"type:varchar(50);unique;not null;comment:用户名"`
	Nickname     string                    `json:"nickname" gorm:"type:varchar(80);not null;comment:昵称"`
	RealName     *string                   `json:"real_name,omitempty" gorm:"type:varchar(20);comment:真实名称"`
	Email        string                    `json:"email,omitempty" gorm:"type:varchar(90)"`
	Gender       foundationenum.UserGender `json:"gender,omitempty" gorm:"type:tinyint(1);comment:性别"`
	Number       *string                   `json:"number,omitempty" gorm:"type:varchar(20);comment:身份标识"`
	Slogan       *string                   `json:"slogan,omitempty" gorm:"type:varchar(50);comment:签名"`
	Organization *string                   `json:"organization,omitempty" gorm:"type:varchar(80);comment:组织"`
	Blog         *string                   `json:"blog,omitempty" gorm:"type:varchar(100);comment:个人主页"`
	QQ           *string                   `json:"qq,omitempty" gorm:"type:varchar(15);comment:QQ"`
	VjudgeId     *string                   `json:"vjudge_id,omitempty" gorm:"type:varchar(15);comment:VjudgeId"`
	Github       *string                   `json:"github,omitempty" gorm:"type:varchar(15);comment:Github"`
	Codeforces   *string                   `json:"codeforces,omitempty" gorm:"type:varchar(20)"`
	CheckInCount int                       `json:"check_in_count,omitempty" gorm:"comment:签到次数"`
	InsertTime   time.Time                 `json:"insert_time" gorm:"type:datetime;not null"`
	ModifyTime   time.Time                 `json:"modify_time" gorm:"type:datetime;not null"`
	Accept       int                       `json:"accept,omitempty"`
	Attempt      int                       `json:"attempt,omitempty"`
	Level        int                       `json:"level,omitempty" gorm:"comment:用户等级"`
	Experience   int                       `json:"experience,omitempty" gorm:"comment:用户经验值"`

	ExperienceUpgrade int `json:"experience_upgrade,omitempty" gorm:"-;comment:当前等级升级所需总经验"`
	ExperienceCurrentLevel     int `json:"experience_current_level,omitempty" gorm:"-;comment:当前等级段已积攒经验"`
}

type UserAccountInfo struct {
	Id       int    `json:"id" gorm:"id"`                       // 数据库索引时真正的Id
	Username string `json:"username" gorm:"username"`           // 对用户展示的唯一标识
	Nickname string `json:"nickname,omitempty" gorm:"nickname"` // 显示的昵称
}

type UserModifyInfo struct {
	Nickname     string                    `json:"nickname" gorm:"type:varchar(80);not null;comment:昵称"`
	RealName     *string                   `json:"real_name,omitempty" gorm:"type:varchar(20);comment:真实名称"`
	Email        string                    `json:"email,omitempty" gorm:"type:varchar(90)"`
	Gender       foundationenum.UserGender `json:"gender,omitempty" gorm:"type:tinyint(1);comment:性别"`
	Number       *string                   `json:"number,omitempty" gorm:"type:varchar(20);comment:身份标识"`
	Slogan       *string                   `json:"slogan,omitempty" gorm:"type:varchar(50);comment:签名"`
	Organization *string                   `json:"organization,omitempty" gorm:"type:varchar(80);comment:组织"`
	Blog         *string                   `json:"blog,omitempty" gorm:"type:varchar(100);comment:个人主页"`
	QQ           *string                   `json:"qq,omitempty" gorm:"type:varchar(15);comment:QQ"`
	VjudgeId     *string                   `json:"vjudge_id,omitempty" gorm:"type:varchar(15);comment:VjudgeId"`
	Github       *string                   `json:"github,omitempty" gorm:"type:varchar(15);comment:Github"`
	Codeforces   *string                   `json:"codeforces,omitempty" gorm:"type:varchar(20)"`
}
