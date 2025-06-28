package foundationmodelmongo

type UserAttemptList struct {
	Count        int      `json:"count" bson:"count"`                 // 次数
	ProblemCount int      `json:"problem_count" bson:"problem_count"` // 去重后的题目数量
	ProblemIds   []string `json:"problem_ids" bson:"problem_ids"`     // 题目ID列表
}

type UserAttemptInfo struct {
	Accept  UserAttemptList `json:"accept" bson:"accept"`   // 接受次数
	Attempt UserAttemptList `json:"attempt" bson:"attempt"` // 尝试次数
}
