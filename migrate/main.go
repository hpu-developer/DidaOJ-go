package main

import (
	foundationflag "foundation/foundation-flag"
	foundationinit "foundation/foundation-init"
	"meta/engine"
	metamysql "meta/meta-mysql"
	metapanic "meta/meta-panic"
	"meta/mongo"
	"meta/subsystem"
	"migrate/application"
	"migrate/config"
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
			mysqlSubsystem := &metamysql.Subsystem{}
			mysqlSubsystem.GetConfig = func() *metamysql.Config {
				return config.GetMysqlConfig()
			}
			return mysqlSubsystem
		},
	)

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			migrateSubsystem := &application.Subsystem{}
			return migrateSubsystem
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
	err = engine.Start(nil, nil, true)
	if err != nil {
		metapanic.ProcessError(err, "engine start error")
		return
	}
}
