package foundationenum

type ProblemAttemptStatus int

var (
	ProblemAttemptStatusNone     ProblemAttemptStatus = 0
	ProblemAttemptStatusAttempt  ProblemAttemptStatus = 1
	ProblemAttemptStatusAccepted ProblemAttemptStatus = 2
)

type ProblemAuth int

var (
	ProblemAuthPublic   ProblemAuth = 0 // 公开
	ProblemAuthPassword ProblemAuth = 1 // 密码，输入密码可以访问
	ProblemAuthPrivate  ProblemAuth = 2 // 私有，指定用户可以访问
)

type ProblemRankType int

const (
	ProblemRankTypeTime       ProblemRankType = 0
	ProblemRankTypeMemory     ProblemRankType = 1
	ProblemRankTypeCodeLength ProblemRankType = 2
)
