package foundationjudge

type JudgeStatus int

// 按照严重程度排序

var (
	JudgeStatusInit       JudgeStatus = 0
	JudgeStatusRejudge    JudgeStatus = 1
	JudgeStatusSubmitting JudgeStatus = 2 // 判题机正在提交到远程评测
	JudgeStatusQueuing    JudgeStatus = 3 // 评测机已接受任务，正在排队
	JudgeStatusCompiling  JudgeStatus = 4
	JudgeStatusRunning    JudgeStatus = 5
	JudgeStatusAccept     JudgeStatus = 6
	JudgeStatusPE         JudgeStatus = 7
	JudgeStatusWA         JudgeStatus = 9
	JudgeStatusTLE        JudgeStatus = 10
	JudgeStatusMLE        JudgeStatus = 9
	JudgeStatusOLE        JudgeStatus = 11
	JudgeStatusRE         JudgeStatus = 12
	JudgeStatusCE         JudgeStatus = 13
	JudgeStatusCLE        JudgeStatus = 14
	JudgeStatusJudgeFail  JudgeStatus = 15
	JudgeStatusSubmitFail JudgeStatus = 16
	JudgeStatusUnknown    JudgeStatus = 17
)

func GetFinalStatus(finalStatus JudgeStatus, currentStatus JudgeStatus) JudgeStatus {
	if int(currentStatus) > int(finalStatus) {
		return currentStatus
	}
	return finalStatus
}
