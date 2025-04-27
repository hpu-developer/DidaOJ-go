package foundationpanic

import (
	"context"
	feishutype "foundation/feishu-type"
	"log/slog"
	metafeishu "meta/meta-feishu"
	metaformat "meta/meta-format"
	metapanic "meta/meta-panic"
	"strings"
)

var GetNotifyAppType func() feishutype.AppType
var GetNoticeGroup func() string

func SendNotifyMessage(format ...any) {
	appType := feishutype.AppTypeMeta
	if GetNotifyAppType != nil {
		appType = GetNotifyAppType()
	}
	appKey := string(appType)
	message := metaformat.Format(format...)
	_, sendErr := metafeishu.GetSubsystem().SendMessageTextToChat(
		context.Background(),
		appKey,
		GetNoticeGroup(),
		message,
	)
	if sendErr != nil {
		slog.Error("send panic message error", "err", sendErr)
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
