package foundationpanic

import (
	foundationconfig "foundation/foundation-config"
	"log/slog"
	metafeishu "meta/meta-feishu"
	metaformat "meta/meta-format"
	metapanic "meta/meta-panic"
	"strings"
)

func SendNotifyMessage(format ...any) {
	message := metaformat.Format(format...)
	err := metafeishu.SendMessageTextToCustomRobot(foundationconfig.GetConfig().Feishu.NotifyRobot, message)
	if err != nil {
		slog.Error("send panic message error", "err", err)
		return
	}
}

func isErrorLog(err error) bool {
	return true
}

func isErrorLogAndNotify(err error) bool {
	if strings.Contains(
		err.Error(),
		"An established connection was aborted by the software in your host machine.",
	) {
		return false
	}
	return true
}

func ProcessErrorCallback(err error, format ...any) {
	if isErrorLog(err) {
		message := metapanic.LogError(err, format...)
		if isErrorLogAndNotify(err) {
			SendNotifyMessage(message)
		}
	}
}

func ProcessPanicCallback(name string, err error, format ...any) {
	message := metapanic.LogPanic(name, err, format...)
	SendNotifyMessage(message)
}
