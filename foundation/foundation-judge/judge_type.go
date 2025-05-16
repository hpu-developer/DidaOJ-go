package foundationjudge

type JudgeType int

var (
	JudgeTypeNormal  JudgeType = 0 // 正常判题（比较输出）
	JudgeTypeSpecial JudgeType = 1 // 特殊判题（由特殊的评测程序判断）
)
