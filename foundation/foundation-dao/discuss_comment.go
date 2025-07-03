package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"time"
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

func (d *DiscussCommentDao) GetCommentEditView(ctx context.Context, id int) (
	*foundationview.DiscussCommentViewEdit,
	error,
) {
	var discussComment foundationview.DiscussCommentViewEdit
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.DiscussComment{}).
		Select("id", "discuss_id", "inserter", "content").
		Where("id = ?", id).
		First(&discussComment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discuss comment error")
	}
	return &discussComment, nil
}

func (d *DiscussCommentDao) GetDiscussCommentList(
	ctx context.Context,
	discussId int,
	page int,
	pageSize int,
) (
	[]*foundationview.DiscussCommentList,
	int,
	error,
) {
	var list []*foundationview.DiscussCommentList
	var total int64
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.DiscussComment{}).
		Where("discuss_id = ?", discussId).
		Count(&total).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count discuss comments, page: %d", page)
	}
	err := d.db.WithContext(ctx).
		Table("discuss_comment AS dc").
		Select(
			`
			dc.*,
			ui.username AS inserter_username,
			ui.nickname AS inserter_nickname,
			um.username AS modifier_username,
			um.nickname AS modifier_nickname
		`,
		).
		Joins("LEFT JOIN user ui ON ui.id = dc.inserter").
		Joins("LEFT JOIN user um ON um.id = dc.modifier").
		Where("dc.discuss_id = ?", discussId).
		Order("dc.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to find discuss comments, page: %d", page)
	}
	return list, int(total), nil
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

func (d *DiscussCommentDao) UpdateContent(
	ctx context.Context,
	userId int,
	id int,
	discussId int,
	content string,
	updateTime time.Time,
) error {
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 更新 Discuss 的 update_time
			res := tx.Model(&foundationmodel.Discuss{}).
				Where("id = ?", discussId).
				Updates(
					map[string]interface{}{
						"updater":     userId,
						"update_time": updateTime,
					},
				)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return metaerror.New("update discuss no document matched, discussId:%d", discussId)
			}

			// 更新 DiscussComment 的 content 和 update_time
			res = tx.Model(&foundationmodel.DiscussComment{}).
				Where("id = ?", id).
				Updates(
					map[string]interface{}{
						"content":     content,
						"modifier":    userId,
						"modify_time": updateTime,
					},
				)
			if res.Error != nil {
				return metaerror.Wrap(res.Error, "update discuss comment content failed, id: %d", id)
			}
			if res.RowsAffected == 0 {
				return metaerror.New("update discuss comment no document matched, id:%d", id)
			}

			return nil
		},
	)
}
