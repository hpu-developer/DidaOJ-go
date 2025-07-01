package foundationmodel

type UserRole struct {
	Id   int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"` // 题目Id
	Name string `json:"name,omitempty" bson:"name,omitempty" gorm:"column:name;size:20"`
}

func (p *UserRole) TableName() string {
	return "user_role"
}

type UserRoleBuilder struct {
	item *UserRole
}

func NewUserRoleBuilder() *UserRoleBuilder {
	return &UserRoleBuilder{
		item: &UserRole{},
	}
}

func (b *UserRoleBuilder) Id(id int) *UserRoleBuilder {
	b.item.Id = id
	return b
}

func (b *UserRoleBuilder) Name(name string) *UserRoleBuilder {
	b.item.Name = name
	return b
}

func (b *UserRoleBuilder) Build() *UserRole {
	return b.item
}
