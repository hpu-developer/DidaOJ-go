package main

import (
	foundationflag "foundation/foundation-flag"
	foundationinit "foundation/foundation-init"
	"meta/engine"
	"meta/meta-http"
	metamogo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/subsystem"
	"web/config"
	"web/router"
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
			mongoSubsystem := &metamogo.Subsystem{}
			mongoSubsystem.GetConfig = func() *metamogo.Config {
				return config.GetMongoConfig()
			}
			return mongoSubsystem
		},
	)

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			httpSubsystem := &metahttp.Subsystem{}
			httpSubsystem.GetPort = func() int32 {
				return config.GetHttpPort()
			}
			metahttp.GetAllowedOrigins = config.GetAllowedOrigins
			httpSubsystem.ProcessGin = router.RegisterRoutes
			return httpSubsystem
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
