package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type ProblemTagDao struct {
	collection *mongo.Collection
	db         *gorm.DB
}

var singletonProblemTagDao = singleton.Singleton[ProblemTagDao]{}

func GetProblemTagDao() *ProblemTagDao {
	return singletonProblemTagDao.GetInstance(
		func() *ProblemTagDao {
			dao := &ProblemTagDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ProblemTagDao) UpdateProblemTags(
	ctx context.Context,
	problemId int,
	tags []int,
) error {
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			// 删除旧的标签
			if err := tx.Model(&foundationmodel.ProblemTag{}).
				Where("id = ?", problemId).
				Delete(&foundationmodel.ProblemTag{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old problem tags")
			}
			if len(tags) > 0 {
				for index, tagId := range tags {
					if err := tx.Create(
						&foundationmodel.ProblemTag{
							Id:    problemId,
							TagId: tagId,
							Index: uint8(index),
						},
					).Error; err != nil {
						return metaerror.Wrap(err, "insert new problem tag")
					}
				}
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}
