package request

import foundationjudge "foundation/foundation-judge"

type JudgeApprove struct {
	ProblemId string                        `json:"problem_id"`
	Language  foundationjudge.JudgeLanguage `json:"language"`
	Code      string                        `json:"code"`
}
