package foundationflag

import (
	"flag"
)

var foundationConfigFile string

func Init() {
	flag.StringVar(
		&foundationConfigFile, "foundation-config", "../foundation/resource/config/foundation.yaml",
		"foundation config file",
	)
	flag.Parse()
}

func GetFoundationConfigFile() string {
	return foundationConfigFile
}
