package request

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
