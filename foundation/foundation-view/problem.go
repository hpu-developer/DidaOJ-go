package foundationview

type ProblemViewTitle struct {
	Id    string `json:"id" bson:"_id"`
	Title string `json:"title" bson:"title"`
}

type ProblemViewAuth struct {
	Id          string `json:"id" bson:"_id"`
	CreatorId   int    `json:"creator_id" bson:"creator_id"`     // 创建者Id
	Private     bool   `json:"private" bson:"private"`           // 是否私有
	Members     []int  `json:"members" bson:"members"`           // 访问权限用户列表，只有在私有题目时才有意义
	AuthMembers []int  `json:"auth_members" bson:"auth_members"` // 题目管理员，对题目有编辑权限
}

type ProblemViewApproveJudge struct {
	Id       string  `json:"id" bson:"_id"`
	OriginOj *string `json:"origin_oj" bson:"origin_oj"` // 题目来源的OJ
	OriginId *string `json:"origin_id" bson:"origin_id"` // 题目来源的Id
}

type ProblemViewAttempt struct {
	Id      string `json:"id" bson:"_id"`
	Accept  int    `json:"accept" bson:"accept"`
	Attempt int    `json:"attempt" bson:"attempt"`
}
