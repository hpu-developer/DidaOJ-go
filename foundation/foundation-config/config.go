package foundationconfig

import (
	feishuchat "foundation/feishu-chat"
	foundationauth "foundation/foundation-auth"
	metafeishu "meta/meta-feishu"
)

type Config struct {
	Auth struct {
		Jwt          string `yaml:"jwt"`           // JWT密钥
		Connect      string `yaml:"connect"`       // 连接
		PasswordSalt string `yaml:"password-salt"` //用户密码盐值
	} `yaml:"auth"`

	Roles map[string][]foundationauth.AuthType `yaml:"roles"` // 角色

	Feishu struct {
		NotifyRobot string                          `yaml:"notify-robot"` // 飞书通知机器人
		App         map[string]metafeishu.AppConfig `yaml:"app"`
		Chat        feishuchat.Config               `yaml:"chat"`
	} `yaml:"feishu"`
}
