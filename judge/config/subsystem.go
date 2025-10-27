package config

import (
	foundationjudge "foundation/foundation-judge"
	cfr2 "meta/cf-r2"
	"meta/engine"
	metaconfig "meta/meta-config"
	metapostgresql "meta/meta-postgresql"
)

type Config struct {
	Judger       foundationjudge.JudgerConfig      `yaml:"judger"`         // 评测器标识
	GoJudge      foundationjudge.GoJudgeConfig     `yaml:"go-judge"`       // GoJudge 数据服务地址
	MaxJob       int                               `yaml:"max-job"`        // 最大同时评测的job数量
	MaxJobRemote int                               `yaml:"max-job-remote"` // 最大同时远程评测的job数量
	JudgeData    cfr2.Config                       `yaml:"judge-data"`     // GoJudge 数据服务地址
	PostgreSql   map[string]*metapostgresql.Config `yaml:"postgresql"`

	CfR2 map[string]*cfr2.Config `yaml:"cf-r2"` // GoJudge 数据服务地址

	Files map[string]string `yaml:"files"`
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

func GetPostgreSqlConfig() map[string]*metapostgresql.Config {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return configSubsystem.config.PostgreSql
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

func GetFilesConfig() map[string]string {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return configSubsystem.config.Files
}

func GetFilesConfigPath(key string) string {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return ""
	}
	if configSubsystem.config == nil {
		return ""
	}
	if configSubsystem.config.Files == nil {
		return ""
	}
	if path, ok := configSubsystem.config.Files[key]; ok {
		return path
	}
	return ""
}
