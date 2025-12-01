package botjudge

import (
	"encoding/json"
	"fmt"
)

func SendError(err error) {
	req := Request{
		Action: ActionTypeError,
		Param:  json.RawMessage(err.Error()),
	}
	fmt.Println(req.Json())
}

func SendLog(log string) {
	req := Request{
		Action: ActionTypeLog,
		Param:  json.RawMessage(log),
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
		return fmt.Errorf("JSON序列化错误: %v", err)
	}

	req := Request{
		Action: ActionTypeInput,
		Param:  json.RawMessage(paramBytes),
	}
	fmt.Println(req.Json())

	return nil
}
