package foundationdao

import (
	"context"
	"errors"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"time"
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

func (d *CollectionDao) CheckUserJoin(ctx context.Context, id int, userId int) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.CollectionMember{}).
		Select("1").
		Where("id = ? AND user_id = ?", id, userId).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, metaerror.Wrap(err, "check user join collection error")
	}
	if dummy > 0 {
		return true, nil
	}
	return false, nil
}

func (d *CollectionDao) GetCollectionDetail(ctx context.Context, id int) (*foundationview.CollectionDetail, error) {
	var result foundationview.CollectionDetail
	err := d.db.WithContext(ctx).
		Table("collection AS c").
		Select(
			`
			c.id, c.title, c.description, c.start_time, c.end_time,
			c.inserter, c.modifier, c.insert_time, c.modify_time, c.password, c.private,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname
		`,
		).
		Joins("LEFT JOIN user AS u1 ON c.inserter = u1.id").
		Joins("LEFT JOIN user AS u2 ON c.modifier = u2.id").
		Where("c.id = ?", id).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, metaerror.Wrap(err, "find collection error")
	}
	return &result, nil
}

func (d *CollectionDao) GetCollectionList(
	ctx context.Context,
	page int,
	pageSize int,
) (
	[]*foundationview.CollectionList,
	int,
	error,
) {
	var list []*foundationview.CollectionList
	var total int64
	offset := (page - 1) * pageSize
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.Collection{}).
		Count(&total).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count collections")
	}
	err := d.db.WithContext(ctx).
		Table("collections AS c").
		Select(
			`
			c.id,
			c.title,
			c.start_time,
			c.end_time,
			c.inserter,
			u.username AS inserter_username,
			u.nickname AS inserter_nickname
		`,
		).
		Joins("LEFT JOIN users u ON u.id = c.inserter").
		Order("c.id DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&list).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to list collections")
	}
	return list, int(total), nil
}

func (d *CollectionDao) GetCollectionRankDetail(
	ctx context.Context,
	collectionId int,
) (*foundationview.CollectionRankDetail, error) {
	var detail foundationview.CollectionRankDetail
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Collection{}).
		Select(`start_time, end_time`).
		Where("id = ?", collectionId).
		Scan(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}
func (d *CollectionDao) GetCollectionRank(
	ctx context.Context,
	collectionId int,
	collection *foundationview.CollectionRankDetail,
) ([]*foundationview.CollectionRank, error) {
	db := d.db.WithContext(ctx).
		Table("judge_job AS j").
		Select(
			`
            j.inserter             AS inserter,
            u.username              AS inserter_username,
            u.nickname              AS inserter_nickname,
            COUNT(CASE WHEN j.status = ? THEN 1 END) AS accept
        `, foundationjudge.JudgeStatusAC,
		).
		Joins("JOIN collection_member AS cm ON cm.user_id = j.inserter AND cm.id = ?", collectionId).
		Joins("JOIN user AS u       ON u.id = j.inserter").
		Where("j.problem_id IN ?", collection.Problems)
	if collection.StartTime != nil {
		db = db.Where("j.insert_time >= ?", *collection.StartTime)
	}
	if collection.EndTime != nil {
		db = db.Where("j.insert_time <= ?", *collection.EndTime)
	}
	db = db.Group("j.inserter")
	var ranks []*foundationview.CollectionRank
	if err := db.Scan(&ranks).Error; err != nil {
		return nil, err
	}
	return ranks, nil
}

func (d *CollectionDao) GetProblemAttemptInfo(
	ctx context.Context,
	collectionId int,
	problemIds []int,
	startTime *time.Time,
	endTime *time.Time,
) ([]*foundationview.ProblemAttemptInfo, error) {
	subMember := d.db.
		Model(&foundationmodel.CollectionMember{}).
		Select("user_id").
		Where("collection_id = ?", collectionId)
	db := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Select(
			"problem_id AS id",
			"COUNT(*) AS attempt",
			"SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS accept", foundationjudge.JudgeStatusAC,
		).
		Where("inserter IN (?)", subMember).
		Where("problem_id IN ?", problemIds)
	if startTime != nil {
		db = db.Where("insert_time >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("insert_time <= ?", *endTime)
	}
	var results []*foundationview.ProblemAttemptInfo
	if err := db.Group("problem_id").Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
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
			collection.Id = 0
			if err := tx.Create(collection).Error; err != nil {
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
