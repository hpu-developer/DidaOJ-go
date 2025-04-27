package foundationinit

import (
	foundationconfig "foundation/foundation-config"
	foundationpanic "foundation/foundation-panic"
	metapanic "meta/meta-panic"
)

func Init() error {
	err := foundationconfig.Init()
	if err != nil {
		return err
	}

	metapanic.ProcessPanicCallback = foundationpanic.ProcessPanicCallback
	metapanic.ProcessErrorCallback = foundationpanic.ProcessErrorCallback

	return nil
}

func Start() error {
	return nil
}
