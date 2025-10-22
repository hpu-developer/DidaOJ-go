package foundationenum

type RemoteJudgeType string

var (
	RemoteJudgeTypeLocal RemoteJudgeType = "DidaOJ"
	RemoteJudgeTypeHdu   RemoteJudgeType = "HDU"
	RemoteJudgeTypePoj   RemoteJudgeType = "POJ"
	RemoteJudgeTypeNyoj  RemoteJudgeType = "NYOJ"
)
