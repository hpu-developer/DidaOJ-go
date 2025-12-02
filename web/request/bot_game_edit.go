package request

import (
	foundationerrorcode "foundation/error-code"
	metaerrorcode "meta/error-code"
)

// BotGameEdit 游戏编辑请求数据结构
type BotGameEdit struct {
	Id          int    `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	JudgeCode   string `json:"judge_code,omitempty"`
}

// CheckRequest 验证请求数据
func (r *BotGameEdit) CheckRequest() (bool, int) {
	if r.Id == 0 {
		return false, int(foundationerrorcode.ParamError)
	}
	if r.Title == "" {
		return false, int(foundationerrorcode.ParamError)
	}
	if r.Description == "" {
		return false, int(foundationerrorcode.ParamError)
	}
	if r.JudgeCode == "" {
		return false, int(foundationerrorcode.ParamError)
	}
	return true, int(metaerrorcode.Success)
}
