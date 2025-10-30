package foundationrequest

type UserModifyInfo struct {
	Nickname string `json:"nickname"`
	Slogan   string `json:"slogan,omitempty"`
}

type UserModifyVjudge struct {
	Approved bool   `json:"approved"`
	Username string `json:"username"`
}
