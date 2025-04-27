package message

import (
	"context"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"meta/meta-feishu"
)

type HandlerQuery struct {
	HandlerOpenId
	IsSelfOpenId func(ctx context.Context, openId string) bool
}

func (h *HandlerQuery) Init() error {
	h.Cmd = "/query"
	return nil
}

func (h *HandlerQuery) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.Command.IsShouldProcess(ctx, event) {
		return false
	}
	if !h.CanProcessChat(ctx, event) {
		return false
	}
	return true
}

func (h *HandlerQuery) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {

	var queryOpenIds []string
	for _, mention := range event.Event.Message.Mentions {
		if h.IsSelfOpenId(ctx, *mention.Id.OpenId) {
			continue
		}
		queryOpenIds = append(queryOpenIds, *mention.Id.OpenId)
	}

	if len(queryOpenIds) == 0 {
		_, err := metafeishu.GetSubsystem().ReplyMessageTextToEvent(ctx, h.AppType, event, "请@所需查询的人")
		if err != nil {
			return false, err
		}
		return false, nil
	}

	for _, openId := range queryOpenIds {
		err := h.ProcessOpenId(ctx, event, openId)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}
