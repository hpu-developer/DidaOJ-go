package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type TagDao struct {
	db *gorm.DB
}

var singletonTagDao = singleton.Singleton[TagDao]{}

func GetTagDao() *TagDao {
	return singletonTagDao.GetInstance(
		func() *TagDao {
			dao := &TagDao{}
			db := metamysql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model(&foundationmodel.Tag{})
			return dao
		},
	)
}

func (d *TagDao) SearchTagIds(ctx context.Context, tag string) ([]int, error) {
	var ids []int
	err := d.db.WithContext(ctx).
		Where("name LIKE ?", "%"+tag+"%").
		Pluck("id", &ids).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to pluck tag ids")
	}
	return ids, nil
}

func (d *TagDao) GetTags(ctx context.Context, ids []int) ([]*foundationmodel.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var tags []*foundationmodel.Tag
	err := d.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&tags).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to get tags by ids")
	}
	return tags, nil
}

func (d *TagDao) InsertTag(ctx context.Context, name string) (int, error) {
	if name == "" {
		return 0, metaerror.New("tag name is empty")
	}
	tag := &foundationmodel.Tag{
		Name: name,
	}
	db := d.db.WithContext(ctx).Model(&foundationmodel.Tag{})
	err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"id": gorm.Expr("LAST_INSERT_ID(id)"),
				},
			),
		},
	).Create(tag).Error
	if err != nil {
		return 0, err
	}
	return tag.Id, nil
}

func (d *TagDao) InsertTagWithDb(tx *gorm.DB, name string) (int, error) {
	if name == "" {
		return 0, metaerror.New("tag name is empty")
	}
	tag := &foundationmodel.Tag{
		Name: name,
	}
	db := tx.Model(&foundationmodel.Tag{})
	err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"id": gorm.Expr("LAST_INSERT_ID(id)"),
				},
			),
		},
	).Create(tag).Error
	if err != nil {
		return 0, err
	}
	return tag.Id, nil
}
