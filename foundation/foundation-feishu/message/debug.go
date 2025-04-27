package message

import (
	"context"
	"encoding/json"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"meta/meta-feishu"
)

type HandlerDebug struct {
	Command
	AppType        feishutype.AppType
	CanProcessChat func(ctx context.Context, event *larkim.P2MessageReceiveV1) bool
}

func (h *HandlerDebug) Init() error {
	h.Cmd = "/debug"
	return nil
}

func (h *HandlerDebug) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.Command.IsShouldProcess(ctx, event) {
		return false
	}
	if !h.CanProcessChat(ctx, event) {
		return false
	}
	return true
}

func (h *HandlerDebug) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {
	message, err := json.Marshal(event)
	if err != nil {
		return false, err
	}
	_, err = metafeishu.GetSubsystem().ReplyMessageTextToEvent(ctx, h.AppType, event, string(message))
	if err != nil {
		return false, err
	}
	return false, nil
}
