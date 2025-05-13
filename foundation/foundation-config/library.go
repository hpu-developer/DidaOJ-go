package foundationconfig

import (
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-flag"
	"gopkg.in/yaml.v3"
	"log/slog"
	metaerror "meta/meta-error"
	metafeishu "meta/meta-feishu"
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

func GetFeishuConfigs() map[string]metafeishu.AppConfig {
	return foundationConfig.Feishu.App
}

func CheckRolesHasAuth(roles []string, auth foundationauth.AuthType) bool {
	if len(roles) == 0 {
		return false
	}
	allRoleAuths := make(map[foundationauth.AuthType]struct{})
	for _, role := range roles {
		auths, ok := GetConfig().Roles[role]
		if !ok {
			continue
		}
		for _, auth := range auths {
			allRoleAuths[auth] = struct{}{}
		}
	}
	if _, ok := allRoleAuths[auth]; !ok {
		return false
	}
	return true
}

func CheckRolesHasAllAuths(roles []string, auths []foundationauth.AuthType) bool {
	if len(auths) == 0 {
		return true
	}
	if len(roles) == 0 {
		return false
	}
	allRoleAuths := make(map[foundationauth.AuthType]struct{})
	for _, role := range roles {
		auths, ok := GetConfig().Roles[role]
		if !ok {
			continue
		}
		for _, auth := range auths {
			allRoleAuths[auth] = struct{}{}
		}
	}
	for _, auth := range auths {
		if _, ok := allRoleAuths[auth]; !ok {
			return false
		}
	}
	return true
}
