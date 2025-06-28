package foundationmodel

type ContestMemberVolunteer struct {
	Id     int `gorm:"column:id;primaryKey"`
	UserId int `gorm:"column:user_id;primaryKey"`
}

func (p *ContestMemberVolunteer) TableName() string {
	return "contest_member_volunteer"
}

type ContestMemberVolunteerBuilder struct {
	item *ContestMemberVolunteer
}

func NewContestMemberVolunteerBuilder() *ContestMemberVolunteerBuilder {
	return &ContestMemberVolunteerBuilder{
		item: &ContestMemberVolunteer{},
	}
}

func (b *ContestMemberVolunteerBuilder) Id(id int) *ContestMemberVolunteerBuilder {
	b.item.Id = id
	return b
}

func (b *ContestMemberVolunteerBuilder) UserId(userId int) *ContestMemberVolunteerBuilder {
	b.item.UserId = userId
	return b
}

func (b *ContestMemberVolunteerBuilder) Build() *ContestMemberVolunteer {
	return b.item
}
