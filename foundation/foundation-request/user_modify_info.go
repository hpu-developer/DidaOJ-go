package foundationrequest

type UserModifyInfo struct {
	Nickname string `json:"nickname"`
	Slogan   string `json:"slogan,omitempty"`
}
