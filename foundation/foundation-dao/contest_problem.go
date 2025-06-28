package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model-mongo"
	"gorm.io/gorm"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type ContestProblemDao struct {
	db *gorm.DB
}

var singletonContestProblemDao = singleton.Singleton[ContestProblemDao]{}

func GetContestProblemDao() *ContestProblemDao {
	return singletonContestProblemDao.GetInstance(
		func() *ContestProblemDao {
			dao := &ContestProblemDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ContestProblemDao) GetProblemIdByContest(ctx context.Context, id int, index int) (int, error) {
	var problemIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestProblem{}).
		Where("contest_id = ? AND `index` = ?", id, index). // index 是保留字，建议加反引号
		Pluck("problem_id", &problemIds).Error
	if err != nil {
		return 0, err
	}
	if len(problemIds) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return problemIds[0], nil
}
