package foundationrequest

type UserModifyInfo struct {
	Nickname     string `json:"nickname"`
	Slogan       string `json:"slogan,omitempty"`
	RealName     string `json:"real_name,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Organization string `json:"organization,omitempty"`
	Blog         string `json:"blog,omitempty"`
}

type UserModifyPassword struct {
	Password    string `json:"password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

type UserModifyUsername struct {
	Username string `json:"username"`
}

type UserModifyVjudge struct {
	Approved bool   `json:"approved"`
	Username string `json:"username"`
}
