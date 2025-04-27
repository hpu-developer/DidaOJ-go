package foundationconfig

import (
	feishuchat "foundation/feishu-chat"
	metafeishu "meta/meta-feishu"
)

type Config struct {
	Auth struct {
		Jwt     string `yaml:"jwt"`     // JWT密钥
		Connect string `yaml:"connect"` // 连接
	} `yaml:"auth"`

	Feishu struct {
		App  map[string]metafeishu.AppConfig `yaml:"app"`
		Chat feishuchat.Config               `yaml:"chat"`
	} `yaml:"feishu"`
}
