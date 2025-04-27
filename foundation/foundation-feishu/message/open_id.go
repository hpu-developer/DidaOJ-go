package message

import (
	"context"
	"encoding/json"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	metaerror "meta/meta-error"
	"meta/meta-feishu"
	"strings"
)

type HandlerOpenId struct {
	Command
	AppType        feishutype.AppType
	CanProcessChat func(ctx context.Context, event *larkim.P2MessageReceiveV1) bool
}

func (h *HandlerOpenId) Init() error {
	h.Cmd = "/open_id"
	return nil
}

func (h *HandlerOpenId) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.Command.IsShouldProcess(ctx, event) {
		return false
	}
	if !h.CanProcessChat(ctx, event) {
		return false
	}
	return true
}

func (h *HandlerOpenId) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {
	openId := strings.TrimSpace(h.RealMessageTextContent)
	err := h.ProcessOpenId(ctx, event, openId)
	if err != nil {
		return false, err
	}
	return false, nil
}

func (h *HandlerOpenId) ProcessOpenId(ctx context.Context, event *larkim.P2MessageReceiveV1, openId string) error {

	feishuClient := metafeishu.GetSubsystem().GetFeishuClient(h.AppType)
	if feishuClient == nil {
		return metaerror.New("feishuClient is nil")
	}

	req := larkcontact.NewGetUserReqBuilder().
		UserId(openId).
		Build()

	resp, err := feishuClient.Contact.User.Get(ctx, req)
	if err != nil {
		return metaerror.Wrap(err)
	}
	if !resp.Success() {
		message := "查询：" + openId + "\n"
		message = message + "查询失败，有可能用户不存在或无权限"
		_, err := metafeishu.GetSubsystem().ReplyMessageTextToEvent(ctx, h.AppType, event, message)
		if err != nil {
			return err
		}
		return nil
	}

	result, err := json.Marshal(resp.Data.User)
	if err != nil {
		return err
	}

	message := "查询：" + openId + "\n"

	message = message + "查询结果：\n" + string(result)

	_, err = metafeishu.GetSubsystem().ReplyMessageTextToEvent(ctx, h.AppType, event, message)
	if err != nil {
		return err
	}

	return nil
}
