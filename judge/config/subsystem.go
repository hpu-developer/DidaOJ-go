package config

import (
	foundationjudge "foundation/foundation-judge"
	cfr2 "meta/cf-r2"
	"meta/engine"
	metaconfig "meta/meta-config"
	metamongo "meta/meta-mongo"
)

type Config struct {
	Judger    string                        `yaml:"judger"`     // 评测器标识
	GoJudge   foundationjudge.GoJudgeConfig `yaml:"go-judge"`   // GoJudge 数据服务地址
	MaxJob    int                           `yaml:"max-job"`    // 最大同时评测的job数量
	JudgeData cfr2.Config                   `yaml:"judge-data"` // GoJudge 数据服务地址
	Mongo     metamongo.Config              `yaml:"mongo"`
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

func GetMongoConfig() *metamongo.Config {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return &configSubsystem.config.Mongo
}
