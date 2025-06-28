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

type ProblemMemberAuthDao struct {
	collection *mongo.Collection
	db         *gorm.DB
}

var singletonProblemMemberAuthDao = singleton.Singleton[ProblemMemberAuthDao]{}

func GetProblemMemberAuthDao() *ProblemMemberAuthDao {
	return singletonProblemMemberAuthDao.GetInstance(
		func() *ProblemMemberAuthDao {
			dao := &ProblemMemberAuthDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ProblemMemberAuthDao) UpdateProblemMemberAuths(
	ctx context.Context,
	problemId int,
	userIds []int,
) error {
	db := d.db
	err := db.Transaction(
		func(tx *gorm.DB) error {
			// 删除旧的标签
			if err := tx.Model(&foundationmodel.ProblemMemberAuth{}).
				Where("id = ?", problemId).
				Delete(&foundationmodel.ProblemMemberAuth{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old problem userIds")
			}
			if len(userIds) > 0 {
				for _, tagId := range userIds {
					if err := tx.Create(
						&foundationmodel.ProblemMemberAuth{
							Id:     problemId,
							UserId: tagId,
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
