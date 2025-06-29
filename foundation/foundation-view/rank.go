package foundationview

type UserRank struct {
	Id           int    `json:"id"`                // 对应 author_id
	Username     string `json:"username"`          // user.username
	Nickname     string `json:"nickname"`          // user.nickname
	Slogan       string `json:"slogan"`            // user.slogan
	ProblemCount int    `json:"problem_count"`     // 统计 count
	Accept       int    `json:"accept,omitempty"`  // AC次数
	Attempt      int    `json:"attempt,omitempty"` // 尝试次数
}
