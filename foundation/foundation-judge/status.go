package foundationjudge

type JudgeStatus int

var (
	JudgeStatusInit      JudgeStatus = 0
	JudgeStatusRejudge   JudgeStatus = 1
	JudgeStatusCompiling JudgeStatus = 2
	JudgeStatusRunning   JudgeStatus = 3
	JudgeStatusAccept    JudgeStatus = 4
	JudgeStatusPE        JudgeStatus = 5
	JudgeStatusWA        JudgeStatus = 6
	JudgeStatusTLE       JudgeStatus = 7
	JudgeStatusMLE       JudgeStatus = 8
	JudgeStatusOLE       JudgeStatus = 9
	JudgeStatusRE        JudgeStatus = 10
	JudgeStatusCE        JudgeStatus = 11
	JudgeStatusCLE       JudgeStatus = 12
	JudgeStatusJudgeFail JudgeStatus = 14
)
