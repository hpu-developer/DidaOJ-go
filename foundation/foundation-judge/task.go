package foundationjudge

type JudgeTaskConfig struct {
	Key      string `json:"key"`       // 任务标识
	Score    int    `json:"score"`     // 代码分数
	InFile   string `json:"in_file"`   // 输入文件
	OutFile  string `json:"out_file"`  // 输出文件
	OutLimit int64  `json:"out_limit"` // 输出长度限制
}

type JudgeJobConfig struct {
	Tasks        []*JudgeTaskConfig `json:"tasks"` // 任务列表
	SpecialJudge *struct {
		Language string `json:"language"` // 程序语言（key）
		Source   string `json:"source"`   // 程序代码
	} `json:"special_judge"` // 特殊判题配置
}
