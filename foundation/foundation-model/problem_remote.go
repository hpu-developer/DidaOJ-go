package foundationmodel

type ProblemRemote struct {
	Id        int `json:"id" gorm:"column:id;primaryKey;autoIncrement"`                          // Remote题目Id
	ProblemId int `json:"problem_id" bson:"problem_id" gorm:"column:problem_id;unique;not null"` // 题目Id

	OriginOj     string  `json:"origin_oj" bson:"origin_oj,omitempty" gorm:"column:origin_oj;size:10;not null"`               // 来源OJ
	OriginId     string  `json:"origin_id" bson:"origin_id,omitempty" gorm:"column:origin_id;size:8;not null"`                // 来源OJ
	OriginUrl    string  `json:"origin_url,omitempty" bson:"origin_url,omitempty" gorm:"column:origin_url;size:100;not null"` // 来源链接
	OriginAuthor *string `json:"origin_author,omitempty" bson:"origin_author,omitempty" gorm:"column:origin_author;size:20"`  // 来源作者
}

func (p *ProblemRemote) TableName() string {
	return "problem_remote"
}

type ProblemRemoteBuilder struct {
	item *ProblemRemote
}

func NewProblemRemoteBuilder() *ProblemRemoteBuilder {
	return &ProblemRemoteBuilder{
		item: &ProblemRemote{},
	}
}

func (b *ProblemRemoteBuilder) Id(id int) *ProblemRemoteBuilder {
	b.item.Id = id
	return b
}

func (b *ProblemRemoteBuilder) ProblemId(problemId int) *ProblemRemoteBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *ProblemRemoteBuilder) OriginOj(originOj string) *ProblemRemoteBuilder {
	b.item.OriginOj = originOj
	return b
}

func (b *ProblemRemoteBuilder) OriginId(originId string) *ProblemRemoteBuilder {
	b.item.OriginId = originId
	return b
}

func (b *ProblemRemoteBuilder) OriginUrl(originUrl string) *ProblemRemoteBuilder {
	b.item.OriginUrl = originUrl
	return b
}

func (b *ProblemRemoteBuilder) OriginAuthor(originAuthor *string) *ProblemRemoteBuilder {
	b.item.OriginAuthor = originAuthor
	return b
}

func (b *ProblemRemoteBuilder) Build() *ProblemRemote {
	return b.item
}
