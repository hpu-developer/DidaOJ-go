package foundationbot

type BotGameStatus int

const (
	BotGameStatusInit      BotGameStatus = 0
	BotGameStatusQueuing   BotGameStatus = iota // 评测机已接受任务，正在排队
	BotGameStatusCompiling BotGameStatus = iota // 评测机正在编译
	BotGameStatusRunning   BotGameStatus = iota
	BotGameStatusFinish    BotGameStatus = iota
	BotGameStatusTLE       BotGameStatus = iota
	BotGameStatusMLE       BotGameStatus = iota
	BotGameStatusOLE       BotGameStatus = iota
	BotGameStatusRE        BotGameStatus = iota
	BotGameStatusCE        BotGameStatus = iota
	BotGameStatusCLE       BotGameStatus = iota
	BotGameStatusJudgeFail BotGameStatus = iota
	BotGameStatusUnknown   BotGameStatus = iota
	BotGameStatusMax       BotGameStatus = iota
)
