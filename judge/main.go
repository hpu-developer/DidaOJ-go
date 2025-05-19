package main

import (
	foundationflag "foundation/foundation-flag"
	foundationinit "foundation/foundation-init"
	"judge/application"
	"judge/config"
	cfr2 "meta/cf-r2"
	"meta/engine"
	"meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/subsystem"
)

func InitPre() error {
	err := foundationinit.Init()
	if err != nil {
		return err
	}

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			return &config.Subsystem{}
		},
	)

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			mongoSubsystem := &metamongo.Subsystem{}
			mongoSubsystem.GetConfig = func() *metamongo.Config {
				return config.GetMongoConfig()
			}
			return mongoSubsystem
		},
	)

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			cfr2Subsystem := &cfr2.Subsystem{}
			cfr2Subsystem.GetConfig = func() map[string]*cfr2.Config {
				return config.GetCfr2Config()
			}
			return cfr2Subsystem
		},
	)

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			judgeSubsystem := &application.Subsystem{}
			return judgeSubsystem
		},
	)

	return nil
}

func main() {
	err := engine.Init(foundationflag.Init, InitPre, nil)
	if err != nil {
		metapanic.ProcessError(err, "engine init error")
		return
	}
	err = engine.Start(nil, nil, false)
	if err != nil {
		metapanic.ProcessError(err, "engine start error")
		return
	}
}
