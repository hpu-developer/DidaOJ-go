package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	"meta/singleton"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagDao struct {
	db *gorm.DB
}

var singletonTagDao = singleton.Singleton[TagDao]{}

func GetTagDao() *TagDao {
	return singletonTagDao.GetInstance(
		func() *TagDao {
			dao := &TagDao{}
			db := metapostgresql.GetSubsystem().GetClient("didaoj")
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

	sql := `
		INSERT INTO tag (name)
		VALUES (?)
		ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)
	`
	if err := tx.Exec(sql, name).Error; err != nil {
		return 0, err
	}
	var id int
	if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&id).Error; err != nil {
		return 0, err
	}
	return id, nil
}
