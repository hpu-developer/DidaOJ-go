package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type DiscussCommentDao struct {
	db *gorm.DB
}

var singletonDiscussCommentDao = singleton.Singleton[DiscussCommentDao]{}

func GetDiscussCommentDao() *DiscussCommentDao {
	return singletonDiscussCommentDao.GetInstance(
		func() *DiscussCommentDao {
			dao := &DiscussCommentDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *DiscussCommentDao) InsertDiscussComment(
	ctx context.Context,
	discussComment *foundationmodel.DiscussComment,
) error {
	if discussComment == nil {
		return metaerror.New("discussComment is nil")
	}

	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(&foundationmodel.Discuss{}).
				Where("id = ?", discussComment.DiscussId).
				Update("updater", discussComment.Inserter).
				Update("update_time", discussComment.InsertTime).Error; err != nil {
				return metaerror.Wrap(err, "update discuss update_time failed")
			}
			if err := tx.Model(&foundationmodel.DiscussComment{}).Create(discussComment).Error; err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					return metaerror.New("discussComment already exists")
				}
				return metaerror.Wrap(err, "insert discussComment failed")
			}
			if discussComment.Id == 0 {
				return metaerror.New("discussComment id is zero after insert")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}
