package foundationdao

import (
	"context"
	"encoding/json"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type ProblemTagDao struct {
	db *gorm.DB
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

func (d *ProblemTagDao) GetProblemTags(ctx context.Context, problemIds []int) ([]int, error) {

	var ids []int
	err := d.db.WithContext(ctx).
		Select("DISTINCT tag_id").
		Where("id IN ?", problemIds).
		Order("index ASC").
		Pluck("tag_id", &ids).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to pluck tag ids")
	}
	return ids, nil
}

func (d *ProblemTagDao) GetProblemTagMap(ctx context.Context, problemIds []int) (map[int][]int, error) {
	type ProblemTagsResult struct {
		ProblemId int             `json:"id" gorm:"column:id"`
		TagIdsRaw json.RawMessage `json:"tag_ids" gorm:"column:tag_ids"`
	}
	var results []ProblemTagsResult
	err := d.db.WithContext(ctx).
		Raw(
			"SELECT id, JSON_ARRAYAGG(tag_id) AS tag_ids "+
				"FROM (SELECT id, tag_id FROM problem_tag WHERE id IN (?) ORDER BY id, `index`) AS sorted_tags "+
				"GROUP BY id",
			problemIds,
		).
		Scan(&results).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to aggregate tag_ids")
	}
	if len(results) == 0 {
		return nil, nil
	}
	tagMap := make(map[int][]int)
	for _, r := range results {
		if r.TagIdsRaw == nil {
			continue
		}
		var tagIds []int
		if err := json.Unmarshal(r.TagIdsRaw, &tagIds); err != nil {
			return nil, metaerror.Wrap(err, "failed to unmarshal tag_ids, problem_id: %d", r.ProblemId)
		}
		tagMap[r.ProblemId] = tagIds
	}
	return tagMap, nil
}

func (d *ProblemTagDao) GetProblemTagList(ctx context.Context, maxCount int) (
	[]*foundationmodel.Tag,
	int,
	error,
) {
	var tags []*foundationmodel.Tag
	var count int64
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Tag{}).
		Where(
			"id IN (?)",
			d.db.Model(&foundationmodel.ProblemTag{}).Select("DISTINCT tag_id"),
		).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = d.db.WithContext(ctx).
		Model(&foundationmodel.Tag{}).
		Where(
			"id IN (?)",
			d.db.Model(&foundationmodel.ProblemTag{}).Select("DISTINCT tag_id"),
		).
		Limit(maxCount).
		Find(&tags).Error
	if err != nil {
		return nil, 0, err
	}
	return tags, int(count), nil
}

func (d *ProblemTagDao) UpdateProblemTags(
	ctx context.Context,
	problemId int,
	tags []int,
) error {
	db := d.db.WithContext(ctx)
	return d.UpdateProblemTagsByDb(db, problemId, tags)
}

func (d *ProblemTagDao) UpdateProblemTagsByDb(
	tx *gorm.DB,
	problemId int,
	tags []int,
) error {
	err := tx.Transaction(
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
