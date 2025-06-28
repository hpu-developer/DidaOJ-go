package foundationenum

type ProblemAttemptStatus int

var (
	ProblemAttemptStatusNone     ProblemAttemptStatus = 0
	ProblemAttemptStatusAttempt  ProblemAttemptStatus = 1
	ProblemAttemptStatusAccepted ProblemAttemptStatus = 2
)
