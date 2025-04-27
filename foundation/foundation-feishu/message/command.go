package message

import (
	"context"
	"encoding/json"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	metamessage "meta/meta-feishu/oapi/message"
	"strings"
)

type Command struct {
	AtOrSingle
	Cmd                    string
	RealMessageTextContent string
}

func (h *Command) IsShouldProcess(ctx context.Context, event *larkim.P2MessageReceiveV1) bool {
	if !h.AtOrSingle.IsShouldProcess(ctx, event) {
		return false
	}
	if !h.CheckProcessCommand(event, h.Cmd) {
		return false
	}
	return true
}

func (h *Command) CheckProcessCommand(event *larkim.P2MessageReceiveV1, command string) bool {
	// 确保消息类型是文本
	if *(event.Event.Message.MessageType) != "text" {
		return false
	}

	// 解析消息内容为 MessageText 结构
	v := metamessage.MessageText{}
	err := json.Unmarshal([]byte(*event.Event.Message.Content), &v)
	if err != nil {
		return false
	}

	// 将消息内容转换为小写
	messageTextContentLower := strings.ToLower(*v.Text)
	commandLower := strings.ToLower(command)

	// 查找命令的位置
	index := strings.Index(messageTextContentLower, commandLower)
	if index < 0 {
		return false
	}

	// 提取命令后的内容
	h.RealMessageTextContent = (*v.Text)[index+len(command):]
	return true
}
