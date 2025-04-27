package message

import (
	"context"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	metafeishu "meta/meta-feishu"
)

type HandlerCommon struct {
	AtOrSingle
	AppType      feishutype.AppType
	IsSelfOpenId func(ctx context.Context, openId string) bool
}

func (h *HandlerCommon) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.AtOrSingle.IsShouldProcess(ctx, event) {
		return false
	}
	senderOpenId := *event.Event.Sender.SenderId.OpenId
	if metafeishu.IsMessageInGroup(event) {
		// 如果在群内同时还@了别人，则本节点不处理
		for _, mention := range event.Event.Message.Mentions {
			openId := *mention.Id.OpenId
			if h.IsSelfOpenId(ctx, openId) {
				continue
			}
			if openId == senderOpenId {
				continue
			}
			return false
		}
	}
	return true
}

func (h *HandlerCommon) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {
	message := "小助手正在工作，欢迎提供相关意见和建议\n部分功能维护中暂不可用，可联系PM催办"
	_, err := metafeishu.GetSubsystem().ReplyMessageTextToEvent(ctx, h.AppType, event, message)
	if err != nil {
		return true, err
	}
	return true, nil
}
