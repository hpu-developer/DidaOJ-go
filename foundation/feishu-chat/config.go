package feishuchat

type Config struct {
	Group struct {
		// 检测错误通知群
		ErrorNotifyGroup string `yaml:"error-notify-group"`
	} `yaml:"group"`
}
