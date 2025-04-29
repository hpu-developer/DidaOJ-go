package gojudge

type Status string

var (
	StatusAccepted      Status = "Accepted"
	StatusMemoryLimit   Status = "Memory Limit Exceeded"
	StatusTimeLimit     Status = "Time Limit Exceeded"
	StatusOutputLimit   Status = "Output Limit Exceeded"
	StatusFileError     Status = "File Error"
	StatusNonzeroExit   Status = "Nonzero Exit Status"
	StatusSignalled     Status = "Signalled"
	StatusInternalError Status = "Internal Error"
)
