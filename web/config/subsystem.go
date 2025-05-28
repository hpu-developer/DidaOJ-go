package config

import (
	foundationjudge "foundation/foundation-judge"
	cfr2 "meta/cf-r2"
	"meta/engine"
	metaconfig "meta/meta-config"
	metaemail "meta/meta-email"
	metamogo "meta/meta-mongo"
	metaredis "meta/meta-redis"
	metastring "meta/meta-string"
)

type Config struct {
	Env struct {
		HttpPort int32 `yaml:"http-port"` // Http端口
	} `yaml:"env"`

	AllowedOrigins []string `yaml:"allowed-origins"` // 允许的跨域请求来源

	Mongo metamogo.Config `yaml:"mongo"`

	Redis metaredis.Config `yaml:"redis"`

	GoJudge foundationjudge.GoJudgeConfig `yaml:"go-judge"` // GoJudge 数据服务地址

	CfTurnstile string `yaml:"cf-turnstile"` // Cloudflare Turnstile 密钥

	CfR2 map[string]*cfr2.Config `yaml:"cf-r2"` // GoJudge 数据服务地址

	Email *metaemail.Config `yaml:"email"`

	Template map[string]string `yaml:"template"`
}

type Subsystem struct {
	metaconfig.Subsystem
	config *Config
}

func GetSubsystem() *Subsystem {
	if thisSubsystem := engine.GetSubsystem[*Subsystem](); thisSubsystem != nil {
		return thisSubsystem.(*Subsystem)
	}
	return nil
}

func (s *Subsystem) Init() error {
	s.config = &Config{}
	return s.InitConfig(s.config)
}

func (s *Subsystem) GetConfig() any {
	return s.config
}

func GetConfig() *Config {
	return GetSubsystem().config
}

func GetHttpPort() int32 {
	if GetSubsystem().config == nil {
		return -1
	}
	return GetSubsystem().config.Env.HttpPort
}

func GetMongoConfig() *metamogo.Config {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return &configSubsystem.config.Mongo
}

func GetAllowedOrigins() []string {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return configSubsystem.config.AllowedOrigins
}

func GetCfr2Config() map[string]*cfr2.Config {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return configSubsystem.config.CfR2
}

func GetRedisConfig() *metaredis.Config {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return &configSubsystem.config.Redis
}

func GetOjTemplateContent(oj string) string {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return ""
	}
	if configSubsystem.config == nil {
		return ""
	}
	if configSubsystem.config.Template == nil {
		return ""
	}
	file, ok := configSubsystem.config.Template[oj]
	if !ok {
		return ""
	}
	content, err := metastring.GetStringFromOpenFile(file)
	if err != nil {
		return ""
	}
	return content
}
