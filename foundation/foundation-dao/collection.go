package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type CollectionDao struct {
	db *gorm.DB
}

var singletonCollectionDao = singleton.Singleton[CollectionDao]{}

func GetCollectionDao() *CollectionDao {
	return singletonCollectionDao.GetInstance(
		func() *CollectionDao {
			dao := &CollectionDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *CollectionDao) UpdateCollection(
	ctx context.Context,
	collection *foundationmodel.Collection,
	problemIds []int,
	members []int,
) error {
	if collection == nil {
		return metaerror.New("collection is nil")
	}
	if len(problemIds) == 0 {
		return metaerror.New("problemIds is empty")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(collection).Where("id = ?", collection.Id).Save(collection).Error; err != nil {
				return metaerror.Wrap(err, "insert collection")
			}
			if err := tx.Model(&foundationmodel.CollectionProblem{}).
				Where("id = ?", collection.Id).
				Delete(&foundationmodel.CollectionProblem{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old problem tags")
			}
			var collectionProblems []*foundationmodel.CollectionProblem
			for i, problemId := range problemIds {
				collectionProblems = append(
					collectionProblems, &foundationmodel.CollectionProblem{
						Id:        collection.Id,
						ProblemId: problemId,
						Index:     i,
					},
				)
			}
			if len(collectionProblems) > 0 {
				if err := tx.Model(&foundationmodel.CollectionProblem{}).
					Create(collectionProblems).Error; err != nil {
					return metaerror.Wrap(err, "insert collection problems")
				}
			}
			if err := tx.Model(&foundationmodel.CollectionMember{}).
				Where("id = ?", collection.Id).
				Delete(&foundationmodel.CollectionMember{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old collection members")
			}
			if len(members) > 0 {
				var collectionMembers []*foundationmodel.CollectionMember
				for _, memberId := range members {
					collectionMembers = append(
						collectionMembers, &foundationmodel.CollectionMember{
							Id:     collection.Id,
							UserId: memberId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.CollectionMember{}).Create(collectionMembers).Error; err != nil {
					return metaerror.Wrap(err, "insert collection members")
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

func (d *CollectionDao) InsertCollection(
	ctx context.Context,
	collection *foundationmodel.Collection,
	problemIds []int,
	members []int,
) error {
	if collection == nil {
		return metaerror.New("collection is nil")
	}
	if len(problemIds) == 0 {
		return metaerror.New("problemIds is empty")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(collection).Create(collection).Error; err != nil {
				return metaerror.Wrap(err, "insert collection")
			}
			if len(problemIds) > 0 {
				var collectionProblems []*foundationmodel.CollectionProblem
				for i, problemId := range problemIds {
					collectionProblems = append(
						collectionProblems, &foundationmodel.CollectionProblem{
							Id:        collection.Id,
							ProblemId: problemId,
							Index:     i,
						},
					)
				}
				if err := tx.Model(&foundationmodel.CollectionProblem{}).
					Create(collectionProblems).Error; err != nil {
					return metaerror.Wrap(err, "insert collection problems")
				}
			}
			if len(members) > 0 {
				var collectionMembers []*foundationmodel.CollectionMember
				for _, memberId := range members {
					collectionMembers = append(
						collectionMembers, &foundationmodel.CollectionMember{
							Id:     collection.Id,
							UserId: memberId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.CollectionMember{}).Create(collectionMembers).Error; err != nil {
					return metaerror.Wrap(err, "insert collection members")
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
