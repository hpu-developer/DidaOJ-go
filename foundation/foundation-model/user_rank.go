package foundationmodel

type UserRank struct {
	Id           int    `json:"id" bson:"_id"`                                          // 数据库索引时真正的Id
	Username     string `json:"username" bson:"username"`                               // 对用户展示的唯一标识
	Nickname     string `json:"nickname,omitempty" bson:"nickname"`                     // 显示的昵称
	ProblemCount int    `json:"problem_count,omitempty" bson:"problem_count,omitempty"` // 解决的题目数
	Accept       int    `json:"accept,omitempty" bson:"accept,omitempty"`               // AC次数
	Attempt      int    `json:"attempt,omitempty" bson:"attempt,omitempty"`             // 尝试次数
}

type UserRankBuilder struct {
	item *UserRank
}

func NewUserRankBuilder() *UserRankBuilder {
	return &UserRankBuilder{item: &UserRank{}}
}

func (b *UserRankBuilder) Id(id int) *UserRankBuilder {
	b.item.Id = id
	return b
}

func (b *UserRankBuilder) Username(username string) *UserRankBuilder {
	b.item.Username = username
	return b
}

func (b *UserRankBuilder) Nickname(nickname string) *UserRankBuilder {
	b.item.Nickname = nickname
	return b
}

func (b *UserRankBuilder) ProblemCount(problemCount int) *UserRankBuilder {
	b.item.ProblemCount = problemCount
	return b
}

func (b *UserRankBuilder) Accept(accept int) *UserRankBuilder {
	b.item.Accept = accept
	return b
}

func (b *UserRankBuilder) Attempt(attempt int) *UserRankBuilder {
	b.item.Attempt = attempt
	return b
}

func (b *UserRankBuilder) Build() *UserRank {
	return b.item
}
