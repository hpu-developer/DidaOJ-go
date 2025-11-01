package foundationrequest

type UserModifyInfo struct {
	Nickname     string `json:"nickname"`
	Slogan       string `json:"slogan,omitempty"`
	RealName     string `json:"real_name,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Organization string `json:"organization,omitempty"`
}

type UserModifyVjudge struct {
	Approved bool   `json:"approved"`
	Username string `json:"username"`
}
