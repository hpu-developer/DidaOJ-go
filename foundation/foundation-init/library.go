package foundationinit

import (
	foundationconfig "foundation/foundation-config"
	foundationpanic "foundation/foundation-panic"
	"meta/engine"
	metafeishu "meta/meta-feishu"
	metapanic "meta/meta-panic"
	"meta/subsystem"
)

func Init() error {
	err := foundationconfig.Init()
	if err != nil {
		return err
	}

	metapanic.ProcessPanicCallback = foundationpanic.ProcessPanicCallback
	metapanic.ProcessErrorCallback = foundationpanic.ProcessErrorCallback

	engine.RegisterSubsystem(
		func() subsystem.Interface {
			feishuSubsystem := &metafeishu.Subsystem{}
			feishuSubsystem.GetFeishuConfigs = foundationconfig.GetFeishuConfigs
			return feishuSubsystem
		},
	)

	return nil
}

func Start() error {
	return nil
}
