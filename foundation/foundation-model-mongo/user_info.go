package foundationmodelmongo

import (
	"time"
)

type UserInfo struct {
	Id           int       `json:"id" bson:"_id"`                                        // 数据库索引时真正的Id
	Username     string    `json:"username" bson:"username"`                             // 对用户展示的唯一标识
	Nickname     string    `json:"nickname,omitempty" bson:"nickname"`                   // 显示的昵称
	Email        string    `json:"email,omitempty" bson:"email,omitempty"`               // 邮箱
	QQ           string    `json:"qq,omitempty" bson:"qq,omitempty"`                     // QQ
	Slogan       string    `json:"slogan,omitempty" bson:"slogan,omitempty"`             // 签名
	Organization string    `json:"organization,omitempty" bson:"organization,omitempty"` // 组织
	RegTime      time.Time `json:"reg_time" bson:"reg_time"`                             // 注册时间
	Accept       int       `json:"accept" bson:"accept"`                                 // AC次数
	Attempt      int       `json:"attempt" bson:"attempt"`                               // 尝试次数
	CheckinCount int       `json:"checkin_count" bson:"checkin_count"`                   // 签到次数

	// 账号关联
	VjudgeId string `json:"vjudge_id,omitempty" bson:"vjudge_id,omitempty"` // vjudge.net Id
}
