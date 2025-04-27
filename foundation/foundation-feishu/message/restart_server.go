package message

import (
	"context"
	"encoding/json"
	"fmt"
	"foundation/enum"
	feishutype "foundation/feishu-type"
	"foundation/foundation-feishu"
	"foundation/text"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	helperErrorCode "helper/error-code"
	helperRequest "helper/request"
	metaerrorcode "meta/error-code"
	metafeishu "meta/meta-feishu"
	metamessage "meta/meta-feishu/oapi/message"
	"sort"
	"strings"
)

type actionData struct {
	Name string
	Type enum.RestartServerType
}

type HandlerServerRestart struct {
	AtOrSingle
	AppType        feishutype.AppType
	IsRestartGroup func(ctx context.Context, event *larkim.P2MessageReceiveV1) bool
	CanOperate     bool

	controllers []string
	actions     []actionData

	serverKey  string
	serverName string
	actionType enum.RestartServerType
}

func (h *HandlerServerRestart) Init() error {
	h.controllers = []string{"更新", "重启"}
	h.actions = []actionData{
		{
			Name: "ds",
			Type: enum.RestartServerTypeDs,
		},
		{
			Name: "",
			Type: enum.RestartServerTypeAll,
		},
	}
	return nil
}

func (h *HandlerServerRestart) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.AtOrSingle.IsShouldProcess(ctx, event) {
		return false
	}
	if *(event.Event.Message.MessageType) != "text" {
		return false
	}
	v := metamessage.MessageText{}
	err := json.Unmarshal([]byte(*event.Event.Message.Content), &v)
	if err != nil {
		return false
	}

	messageTextContentLower := strings.ToLower(*v.Text)

	servers, err := foundationservice.GetServerService().GetServers(ctx)
	if err != nil {
		return false
	}

	// 根据文本长度排序，避免先匹配到了短的，而没有匹配到长的
	var sortServers []string
	serverNames := map[string]string{}
	for _, server := range *servers {
		if server.NameShort == nil {
			continue
		}
		sortServers = append(sortServers, *server.NameShort)
		serverNames[*server.NameShort] = *server.Name
	}
	sort.Slice(
		sortServers, func(i, j int) bool {
			return len(sortServers[i]) > len(sortServers[j])
		},
	)
	for _, controller := range h.controllers {
		for _, action := range h.actions {
			for _, server := range sortServers {
				command := controller + server + action.Name
				if strings.Contains(messageTextContentLower, command) {
					h.serverKey = server
					h.actionType = action.Type
					h.serverName = serverNames[server]
					return true
				}
			}
		}
	}

	return false
}

func (h *HandlerServerRestart) DoProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) (bool, error) {

	if !h.CanOperate {
		message := fmt.Sprintf("识别到需要更新%s\n", h.serverName)
		message = message + fmt.Sprintf("本助手无法直接触发服务器更新")
		message = message +
			fmt.Sprintf(
				"请使用[StarsHelper](https://starshelper.yingxiong.com/server)或者%s",
				metafeishu.GetFeishuMessageTextByFeishuOpenId(foundationfeishu.GetFeishuOpenIdBySuiren(h.AppType)),
			)
		_, sendErr := metafeishu.GetSubsystem().ReplyMessageTextToEvent(
			ctx,
			h.AppType,
			event,
			message,
		)
		if sendErr != nil {
			return false, sendErr
		}
		return false, nil
	}

	messageText := metamessage.MessageText{}
	err := json.Unmarshal([]byte(*event.Event.Message.Content), &messageText)
	if err != nil {
		return false, err
	}

	openId := event.Event.Sender.SenderId.OpenId
	developer, err := foundationservice.GetDeveloperService().GetDeveloperByFeishuOpenId(ctx, h.AppType, *openId)
	if err != nil {
		message := fmt.Sprintf("识别到需要更新%s\n", h.serverName)
		message = message + "小助手中未注册您的信息，可登陆[StarsHelper](https://starshelper.yingxiong.com/login)自动完成注册后重试"
		_, sendErr := metafeishu.GetSubsystem().ReplyMessageTextToEvent(
			ctx, h.AppType,
			event,
			message,
		)
		if sendErr != nil {
			return false, sendErr
		}
		return false, err
	}

	reason := *messageText.Text
	reasonMessage := *messageText.Text

	for _, mention := range event.Event.Message.Mentions {
		reason = strings.ReplaceAll(reason, *mention.Key, fmt.Sprintf("@%s", *mention.Name))
		reasonMessage = strings.ReplaceAll(
			reasonMessage,
			*mention.Key,
			metafeishu.GetFeishuMessageByFeishuOpenId(*mention.Id.OpenId),
		)
	}

	requestData := helperRequest.NewServerRestartApplyRequestBuilder().
		Server(h.serverKey).
		Type(h.actionType).
		Reason(reason).
		ReasonMessage(reasonMessage).
		IsSilent(false).
		Build()

	_, err = foundationservice.GetServerService().ApplyRestart(
		ctx,
		developer,
		enum.RestartServerDataSourceSuiren,
		requestData,
	)
	if err != nil {
		code := metaerrorcode.CommonError
		isDuplicateKeyError := mongo.IsDuplicateKeyError(pkgerrors.Cause(err))
		if isDuplicateKeyError {
			code = helperErrorCode.ServerRestartApplyDuplicate
		}
		message := fmt.Sprintf("识别到需要更新%s，申请失败\n%s", h.serverName, text.GetTextZhByCode(ctx, code))
		_, sendErr := metafeishu.GetSubsystem().ReplyMessageTextToEvent(
			ctx,
			h.AppType,
			event,
			message,
		)
		if sendErr != nil {
			return false, sendErr
		}
		// 如果是因为已经有同类任务，则不认为是逻辑错误
		if isDuplicateKeyError {
			return false, nil
		}
		return false, err
	}

	if h.IsRestartGroup(ctx, event) {
		return false, nil
	}

	message := fmt.Sprintf(
		"识别到%s需要%s\n正在触发，请关注群消息",
		h.serverName,
		enum.GetRestartServerTypeDescription(h.actionType),
	)
	_, err = metafeishu.GetSubsystem().ReplyMessageTextToEvent(
		ctx,
		h.AppType,
		event,
		message,
	)
	if err != nil {
		return false, err
	}
	return false, nil
}
