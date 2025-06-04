package request

import foundationjudge "foundation/foundation-judge"

type JudgeApprove struct {
	ProblemId    string                        `json:"problem_id"`
	ContestId    int                           `json:"contest_id"`
	ProblemIndex int                           `json:"problem_index"`
	Language     foundationjudge.JudgeLanguage `json:"language"`
	Code         string                        `json:"code"`
}
