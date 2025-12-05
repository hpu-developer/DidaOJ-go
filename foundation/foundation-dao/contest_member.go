package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
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
