package foundationconfig

import (
	"foundation/foundation-flag"
	"gopkg.in/yaml.v3"
	"log/slog"
	metaerror "meta/meta-error"
	"os"
)

var foundationConfig Config

func Init() error {
	configFile := foundationflag.GetFoundationConfigFile()
	slog.Info("foundation Config init", "configFile", configFile)
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &foundationConfig)
	if err != nil {
		return metaerror.Wrap(err, "Unmarshal config file error")
	}
	slog.Info("foundation Config", "foundationConfig", foundationConfig)
	return nil
}

func GetConfig() *Config {
	return &foundationConfig
}

func GetJwtSecret() []byte {
	return []byte(foundationConfig.Auth.Jwt)
}
