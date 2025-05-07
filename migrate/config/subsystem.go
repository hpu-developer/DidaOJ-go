package config

import (
	"meta/engine"
	metaconfig "meta/meta-config"
	metamongo "meta/meta-mongo"
	metamysql "meta/meta-mysql"
)

type Config struct {
	OnlyInit bool                         `yaml:"only-init"` // 仅初始化，不导入
	Mongo    metamongo.Config             `yaml:"mongo"`
	Mysql    map[string]*metamysql.Config `yaml:"mysql"`
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

func GetMysqlConfig() map[string]*metamysql.Config {
	configSubsystem := GetSubsystem()
	if configSubsystem == nil {
		return nil
	}
	if configSubsystem.config == nil {
		return nil
	}
	return configSubsystem.config.Mysql
}
