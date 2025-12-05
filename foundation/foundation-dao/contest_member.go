package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContestMemberDao struct {
	db *gorm.DB
}

var singletonContestMemberDao = singleton.Singleton[ContestMemberDao]{}

func GetContestMemberDao() *ContestMemberDao {
	return singletonContestMemberDao.GetInstance(
		func() *ContestMemberDao {
			dao := &ContestMemberDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ContestMemberDao) GetUserIds(ctx context.Context, id int) (
	[]int,
	error,
) {
	var userIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestMember{}).
		Where("id = ?", id).
		Pluck("user_id", &userIds).Error
	if err != nil {
		return nil, err
	}
	if len(userIds) == 0 {
		return nil, nil
	}
	return userIds, nil
}

func (d *ContestMemberDao) GetUser(ctx context.Context, id int, userId int) (
	*foundationview.ContestMember,
	error,
) {
	var user foundationview.ContestMember
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestMember{}).
		Select("user_id as id", "contest_name").
		Where("id = ?", id).
		Where("user_id = ?", userId).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *ContestMemberDao) GetUsersWithName(ctx context.Context, id int) (
	[]*foundationview.ContestMember,
	error,
) {
	var users []*foundationview.ContestMember
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestMember{}).
		Select("user_id as id", "contest_name").
		Where("id = ?", id).
		Where("contest_name is not null and contest_name != ''").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users, nil
}

func (d *ContestMemberDao) PostContestMemberName(
	ctx context.Context, userId int, contestId int, name string,
) error {
	return d.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id"},
				{Name: "user_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"contest_name": name,
			}),
		}).
		Create(&foundationmodel.ContestMember{
			Id:          contestId,
			UserId:      userId,
			ContestName: name,
		}).Error
}
