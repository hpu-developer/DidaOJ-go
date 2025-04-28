package main

import (
	foundationflag "foundation/foundation-flag"
	foundationinit "foundation/foundation-init"
	"judge/application"
	"judge/config"
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
			mongoSubsystem := &mongo.Subsystem{}
			mongoSubsystem.GetConfig = func() *mongo.Config {
				return config.GetMongoConfig()
			}
			return mongoSubsystem
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
