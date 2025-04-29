package config

import (
	"meta/engine"
	metaconfig "meta/meta-config"
	metamogo "meta/meta-mongo"
)

type Config struct {
	Env struct {
		HttpPort int32 `yaml:"http-port"` // Http端口
	} `yaml:"env"`

	Mongo metamogo.Config `yaml:"mongo"`
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
