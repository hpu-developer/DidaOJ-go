package botjudge

import (
	"encoding/json"
)

type ActionType int

const (
	ActionNone            ActionType = 0
	ActionTypeLog         ActionType = 1 // 用于输出日志，一般用于调试
	ActionTypeError       ActionType = 2 // 用于输出报错，一般收到后认为判题异常
	ActionTypeAgentInput  ActionType = 3 // 用于传递Agent的输入
	ActionTypeAgentOutput ActionType = 4 // 用于传递Agent的输出
	ActionTypeInfo        ActionType = 5 // 输出Info信息
	ActionTypeParam       ActionType = 6 // 输出Param信息
	ActionTypeFinish      ActionType = 7 // 表示结束可以关闭
)

type Request struct {
	Action ActionType      `json:"action"`
	Param  json.RawMessage `json:"param"`
}

func (r *Request) Json() string {
	json, _ := json.Marshal(r)
	return string(json)
}
