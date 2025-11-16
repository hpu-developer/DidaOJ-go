package request

import (
	foundationjudge "foundation/foundation-judge"
)

type RunCode struct {
	Language foundationjudge.JudgeLanguage `json:"language" binding:"required"` // 编程语言
	Code     string                        `json:"code" binding:"required"`     // 代码
	Input    string                        `json:"input,omitempty"`             // 输入数据
}
