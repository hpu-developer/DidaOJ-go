package foundationjudge

type JudgeTaskConfig struct {
	Key     string `json:"key"`      // 任务标识
	Score   int    `json:"score"`    // 代码分数
	InFile  string `json:"in_file"`  // 输入文件
	OutFile string `json:"out_file"` // 输出文件
}
