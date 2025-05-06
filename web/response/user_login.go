package response

type UserLogin struct {
	Token    string   `json:"token"`
	UserId   int      `json:"user_id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Roles    []string `json:"roles,omitempty"`
}

type UserLoginBuilder struct {
	item *UserLogin
}

func NewUserLoginBuilder() *UserLoginBuilder {
	return &UserLoginBuilder{
		item: &UserLogin{},
	}
}

func (b *UserLoginBuilder) Token(token string) *UserLoginBuilder {
	b.item.Token = token
	return b
}

func (b *UserLoginBuilder) UserId(userId int) *UserLoginBuilder {
	b.item.UserId = userId
	return b
}

func (b *UserLoginBuilder) Username(username string) *UserLoginBuilder {
	b.item.Username = username
	return b
}

func (b *UserLoginBuilder) Nickname(nickname string) *UserLoginBuilder {
	b.item.Nickname = nickname
	return b
}

func (b *UserLoginBuilder) Roles(roles []string) *UserLoginBuilder {
	b.item.Roles = roles
	return b
}

func (b *UserLoginBuilder) Build() *UserLogin {
	return b.item
}
