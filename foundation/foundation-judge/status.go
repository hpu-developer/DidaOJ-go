package foundationjudge

type JudgeStatus int

// 按照严重程度排序

const (
	JudgeStatusInit       JudgeStatus = 0
	JudgeStatusRejudge    JudgeStatus = 1
	JudgeStatusSubmitting JudgeStatus = 2 // 判题机正在提交到远程评测
	JudgeStatusQueuing    JudgeStatus = 3 // 评测机已接受任务，正在排队
	JudgeStatusCompiling  JudgeStatus = 4
	JudgeStatusRunning    JudgeStatus = 5
	JudgeStatusAC         JudgeStatus = 6
	JudgeStatusPE         JudgeStatus = 7
	JudgeStatusWA         JudgeStatus = 8
	JudgeStatusTLE        JudgeStatus = 9
	JudgeStatusMLE        JudgeStatus = 10
	JudgeStatusOLE        JudgeStatus = 11
	JudgeStatusRE         JudgeStatus = 12
	JudgeStatusCE         JudgeStatus = 13
	JudgeStatusCLE        JudgeStatus = 14
	JudgeStatusJudgeFail  JudgeStatus = 15
	JudgeStatusSubmitFail JudgeStatus = 16
	JudgeStatusUnknown    JudgeStatus = 17
	JudgeStatusMax        JudgeStatus = iota
)

func GetFinalStatus(finalStatus JudgeStatus, currentStatus JudgeStatus) JudgeStatus {
	if int(currentStatus) > int(finalStatus) {
		return currentStatus
	}
	return finalStatus
}

func IsValidJudgeStatus(status int) bool {
	return status >= int(JudgeStatusInit) && status < int(JudgeStatusMax)
}

func IsJudgeStatusRunning(status JudgeStatus) bool {
	switch status {
	case JudgeStatusInit:
	case JudgeStatusRejudge:
	case JudgeStatusSubmitting:
	case JudgeStatusQueuing:
	case JudgeStatusCompiling:
	case JudgeStatusRunning:
		return true
	default:
		return false
	}
	return false
}
