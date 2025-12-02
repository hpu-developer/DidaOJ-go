package botjudge

import (
	"encoding/json"
	"fmt"
)

type Content struct {
	Content string `json:"content"`
}

func SendError(err error) {
	content := Content{
		Content: err.Error(),
	}
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		fmt.Printf("JSON marshal error: %s", err.Error())
		return
	}
	req := Request{
		Action: ActionTypeError,
		Param:  json.RawMessage(jsonBytes),
	}
	fmt.Println(req.Json())
}

func SendLog(log string) {
	content := Content{
		Content: log,
	}
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		SendError(fmt.Errorf("JSON marshal error: %v", err))
		return
	}
	req := Request{
		Action: ActionTypeLog,
		Param:  json.RawMessage(jsonBytes),
	}
	fmt.Println(req.Json())
}

func SendInput(index int, input string) error {
	param := ChannelContent{
		Index:   index,
		Content: input,
	}
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %v", err)
	}

	req := Request{
		Action: ActionTypeAgentInput,
		Param:  json.RawMessage(paramBytes),
	}
	fmt.Println(req.Json())

	return nil
}

func SendInfo(info interface{}) error {
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}
	req := Request{
		Action: ActionTypeInfo,
		Param:  json.RawMessage(jsonBytes),
	}
	fmt.Println(req.Json())
	return nil
}

func SendFinish() {
	req := Request{
		Action: ActionTypeFinish,
	}
	fmt.Println(req.Json())
}
