package foundationrequest

type UserModifyInfo struct {
	Nickname string `json:"nickname"`
	Slogan   string `json:"slogan,omitempty"`
}

type UserModifyVjudge struct {
	Username string `json:"username"`
}
