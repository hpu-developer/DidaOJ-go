package foundationconfig

type Config struct {
	Auth struct {
		Jwt     string `yaml:"jwt"`     // JWT密钥
		Connect string `yaml:"connect"` // 连接
	} `yaml:"auth"`
}
