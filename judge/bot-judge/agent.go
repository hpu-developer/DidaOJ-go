package botjudge

type AgentInfo struct {
	Id       int    `json:"id"`
	Version  int    `json:"version"`
	Nickname string `json:"nickname"`
}

type JudgeInfo struct {
	Agents []AgentInfo `json:"agents"`
}
