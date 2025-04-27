package message

import (
	"context"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type AtOrSingle struct {
	Handler
	IsAtMe func(ctx context.Context, event *larkim.P2MessageReceiveV1) bool
}

func (h *AtOrSingle) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if event.Event.Message.ChatType != nil && *event.Event.Message.ChatType == "group" {
		return h.IsAtMe(ctx, event)
	} else {
		return true
	}
}
