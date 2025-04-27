package message

import (
	"context"
	"fmt"
	foundationconfig "foundation/foundation-config"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	metafeishu "meta/meta-feishu"
)

type HandlerAtOther struct {
	Commands
	AppType      feishutype.AppType
	IsSelfOpenId func(ctx context.Context, openId string) bool
}

func (h *HandlerAtOther) isEnableGroup(chatId string) bool {
	chatGroup := foundationconfig.GetConfig().Chat.Group
	return chatId == chatGroup.WatcherGroup ||
		chatId == chatGroup.DataTableGroup ||
		chatId == chatGroup.RestartServerGroup
}

func (h *HandlerAtOther) Init() error {
	h.CmdList = []string{"/invite", "/notify", "/at"}
	return nil
}

func (h *HandlerAtOther) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.AtOrSingle.IsShouldProcess(ctx, event) {
		return false
	}
	if !metafeishu.IsMessageInGroup(event) {
		return false
	}
	chatId := *event.Event.Message.ChatId
	if !h.isEnableGroup(chatId) {
		if !h.Commands.IsShouldProcess(ctx, event) {
			return false
		}
	}
	return true
}

func (h *HandlerAtOther) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {
	senderOpenId := *event.Event.Sender.SenderId.OpenId
	otherOpenIds := make([]string, 0)
	for _, mention := range event.Event.Message.Mentions {
		openId := *mention.Id.OpenId
		if h.IsSelfOpenId(ctx, openId) {
			continue
		}
		if openId == senderOpenId {
			continue
		}
		otherOpenIds = append(otherOpenIds, openId)
	}
	if len(otherOpenIds) < 1 {
		return true, nil
	}
	feishuSubsystem := metafeishu.GetSubsystem()

	processSuccessAny := false

	for _, openId := range otherOpenIds {
		message := fmt.Sprintf(
			"%s在以下群中@了您，请前往查看",
			metafeishu.GetFeishuMessageTextByFeishuOpenId(senderOpenId),
		)
		_, err := feishuSubsystem.SendMessageTextToOpenId(
			ctx,
			h.AppType,
			openId,
			message,
		)
		if err != nil {
			if metafeishu.IsFeishuErrorNoAvailabilityToUser(err) {
				continue
			}
			return false, err
		}
		_, err = feishuSubsystem.SendMessageChatShareToOpenId(ctx, h.AppType, openId, *event.Event.Message.ChatId)
		if err != nil {
			return false, err
		}
		_, err = feishuSubsystem.MergeForwardMessageToOpenId(ctx, h.AppType, openId, *event.Event.Message.MessageId)
		if err != nil {
			return false, err
		}
		err = feishuSubsystem.InviteUserToChat(ctx, h.AppType, *event.Event.Message.ChatId, openId)
		if err != nil {
			return false, err
		}

		processSuccessAny = true
	}

	if !processSuccessAny {
		return true, nil
	}

	var message string
	for _, openId := range otherOpenIds {
		message = message + metafeishu.GetFeishuMessageTextByFeishuOpenId(openId)
	}
	message = message + "已私聊提醒"
	_, err := feishuSubsystem.ReplyMessageText(
		ctx,
		h.AppType,
		*event.Event.Message.MessageId,
		message,
	)
	if err != nil {
		return false, err
	}

	return false, nil
}
