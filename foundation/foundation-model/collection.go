package foundationmodel

import (
	"time"
)

type Collection struct {
	Id            int        `json:"id" bson:"_id"`                                            // 数据库索引时真正的Id
	Title         string     `json:"title" bson:"title"`                                       // 题集标题
	Description   string     `json:"description,omitempty" bson:"description,omitempty"`       // 题集描述
	StartTime     *time.Time `json:"start_time" bson:"start_time"`                             // 题集开始时间
	EndTime       *time.Time `json:"end_time" bson:"end_time"`                                 // 题集结束时间
	OwnerId       int        `json:"owner_id" bson:"owner_id"`                                 // 题集组织者
	OwnerUsername *string    `json:"owner_username,omitempty" bson:"owner_username,omitempty"` // 题集组织者用户名
	OwnerNickname *string    `json:"owner_nickname,omitempty" bson:"owner_nickname,omitempty"` // 题集组织者昵称

	CreateTime time.Time `json:"create_time" bson:"create_time"` // 创建时间
	UpdateTime time.Time `json:"update_time" bson:"update_time"` // 更新时间

	// 权限相关
	Auth     ContestAuth `json:"auth" bson:"auth"`                             // 题集权限
	Password *string     `json:"password,omitempty" bson:"password,omitempty"` // 题集密码
	Members  []int       `json:"members,omitempty" bson:"members,omitempty"`   // 题集成员，用于控制访问权限与展示排名

	// 题目相关
	Problems []string `json:"problems,omitempty" bson:"problems,omitempty"` // 题目Id列表
}

type CollectionBuilder struct {
	item *Collection
}

func NewCollectionBuilder() *CollectionBuilder {
	return &CollectionBuilder{item: &Collection{}}
}

func (b *CollectionBuilder) Id(id int) *CollectionBuilder {
	b.item.Id = id
	return b
}

func (b *CollectionBuilder) Title(title string) *CollectionBuilder {
	b.item.Title = title
	return b
}

func (b *CollectionBuilder) Description(description string) *CollectionBuilder {
	b.item.Description = description
	return b
}

func (b *CollectionBuilder) StartTime(startTime *time.Time) *CollectionBuilder {
	b.item.StartTime = startTime
	return b
}

func (b *CollectionBuilder) EndTime(endTime *time.Time) *CollectionBuilder {
	b.item.EndTime = endTime
	return b
}

func (b *CollectionBuilder) OwnerId(ownerId int) *CollectionBuilder {
	b.item.OwnerId = ownerId
	return b
}

func (b *CollectionBuilder) OwnerUsername(ownerUsername *string) *CollectionBuilder {
	b.item.OwnerUsername = ownerUsername
	return b
}

func (b *CollectionBuilder) OwnerNickname(ownerNickname *string) *CollectionBuilder {
	b.item.OwnerNickname = ownerNickname
	return b
}

func (b *CollectionBuilder) CreateTime(createTime time.Time) *CollectionBuilder {
	b.item.CreateTime = createTime
	return b
}

func (b *CollectionBuilder) UpdateTime(updateTime time.Time) *CollectionBuilder {
	b.item.UpdateTime = updateTime
	return b
}

func (b *CollectionBuilder) Auth(auth ContestAuth) *CollectionBuilder {
	b.item.Auth = auth
	return b
}

func (b *CollectionBuilder) Password(password *string) *CollectionBuilder {
	b.item.Password = password
	return b
}

func (b *CollectionBuilder) Problems(problems []string) *CollectionBuilder {
	b.item.Problems = problems
	return b
}

func (b *CollectionBuilder) Members(members []int) *CollectionBuilder {
	b.item.Members = members
	return b
}

func (b *CollectionBuilder) Build() *Collection {
	return b.item
}
