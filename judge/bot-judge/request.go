package botjudge

import (
	"encoding/json"
)

type ActionType int

const (
	ActionNone            ActionType = 0
	ActionTypeLog         ActionType = 1
	ActionTypeError       ActionType = 2
	ActionTypeAgentInput  ActionType = 3
	ActionTypeAgentOutput ActionType = 4
)

type Request struct {
	Action ActionType      `json:"action"`
	Param  json.RawMessage `json:"param"`
}

func (r *Request) Json() string {
	json, _ := json.Marshal(r)
	return string(json)
}
