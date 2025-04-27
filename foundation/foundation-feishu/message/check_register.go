package message

import (
	"context"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"meta/meta-feishu"
)

type HandlerCheckRegister struct {
	AtOrSingle
	AppType feishutype.AppType
}

func (h *HandlerCheckRegister) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {
	senderOpenId := event.Event.Sender.SenderId.OpenId
	_, err := foundationdao.GetDeveloperDao().GetDeveloperByFeishuOpenId(ctx, h.AppType, *senderOpenId)
	if err != nil {
		message := "小助手中未注册您的信息，可登陆[StarsHelper](https://starshelper.yingxiong.com/login)自动完成注册"
		_, err := metafeishu.GetSubsystem().ReplyMessageTextToEvent(ctx, h.AppType, event, message)
		if err != nil {
			return false, err
		}
		return false, err
	}
	return true, nil
}
