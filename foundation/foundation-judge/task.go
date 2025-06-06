package foundationjudge

type JudgeTaskConfig struct {
	Key      string `json:"key"`                                          // 任务标识
	Score    int    `json:"score"`                                        // 代码分数
	InFile   string `json:"in_file,omitempty" yaml:"in-file,omitempty"`   // 输入文件
	OutFile  string `json:"out_file,omitempty" yaml:"out-file,omitempty"` // 输出文件
	OutLimit int64  `json:"out_limit" yaml:"out-limit"`                   // 输出长度限制
}

type SpecialJudgeConfig struct {
	Language string `json:"language" yaml:"language"` // 程序语言（key）
	Source   string `json:"source" yaml:"source"`     // 程序代码
}

type JudgeJobConfig struct {
	Tasks        []*JudgeTaskConfig  `json:"tasks"`                                                  // 任务列表
	SpecialJudge *SpecialJudgeConfig `json:"special_judge,omitempty" yaml:"special-judge,omitempty"` // 特判
}
