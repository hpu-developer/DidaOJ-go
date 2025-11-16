package foundationrun

type RunStatus int

// 按照严重程度排序

const (
	RunStatusInit      RunStatus = 0
	RunStatusQueuing   RunStatus = 1 // 评测机已接受任务，正在排队
	RunStatusCompiling RunStatus = 2 // 评测机正在编译代码
	RunStatusRunning   RunStatus = 3
	RunStatusFinish    RunStatus = 4
	RunStatusTLE       RunStatus = 5
	RunStatusMLE       RunStatus = 6
	RunStatusOLE       RunStatus = 7
	RunStatusRE        RunStatus = 8
	RunStatusCE        RunStatus = 9
	RunStatusCLE       RunStatus = 10
	RunStatusRunFail   RunStatus = 11
	RunStatusMax       RunStatus = iota
)

func GetFinalStatus(finalStatus RunStatus, currentStatus RunStatus) RunStatus {
	if int(currentStatus) > int(finalStatus) {
		return currentStatus
	}
	return finalStatus
}

func IsValidRunStatus(status int) bool {
	return status >= int(RunStatusInit) && status < int(RunStatusMax)
}

func IsRunStatusRunning(status RunStatus) bool {
	switch status {
	case RunStatusInit:
		fallthrough
	case RunStatusQueuing:
		fallthrough
	case RunStatusCompiling:
		fallthrough
	case RunStatusRunning:
		return true
	default:
		return false
	}
}
