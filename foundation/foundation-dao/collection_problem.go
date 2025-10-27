package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
)

type CollectionProblemDao struct {
	db *gorm.DB
}

var singletonCollectionProblemDao = singleton.Singleton[CollectionProblemDao]{}

func GetCollectionProblemDao() *CollectionProblemDao {
	return singletonCollectionProblemDao.GetInstance(
		func() *CollectionProblemDao {
			dao := &CollectionProblemDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *CollectionProblemDao) GetCollectionProblems(ctx context.Context, id int) ([]int, error) {
	var problems []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.CollectionProblem{}).
		Where("id = ?", id).
		Order("index ASC"). // 加反引号防止关键词冲突
		Pluck("problem_id", &problems).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get collection problems error")
	}
	return problems, nil
}
