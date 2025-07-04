package foundationdaomongo

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model-mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"time"
)

type DiscussCommentDao struct {
	collection *mongo.Collection
}

var singletonDiscussCommentDao = singleton.Singleton[DiscussCommentDao]{}

func GetDiscussCommentDao() *DiscussCommentDao {
	return singletonDiscussCommentDao.GetInstance(
		func() *DiscussCommentDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var DiscussCommentDao DiscussCommentDao
			DiscussCommentDao.collection = client.
				Database("didaoj").
				Collection("discuss_comment")
			return &DiscussCommentDao
		},
	)
}

func (d *DiscussCommentDao) InitDao(ctx context.Context) error {
	return nil
}

func (d *DiscussCommentDao) GetListAll(ctx context.Context) ([]*foundationmodel.DiscussComment, error) {
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := d.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find all DiscussComments error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var contests []*foundationmodel.DiscussComment
	for cursor.Next(ctx) {
		var contest foundationmodel.DiscussComment
		if err := cursor.Decode(&contest); err != nil {
			return nil, metaerror.Wrap(err, "decode DiscussComment error")
		}
		contests = append(contests, &contest)
	}
	return contests, nil
}

func (d *DiscussCommentDao) GetCommentEditView(ctx *gin.Context, id int) (
	*foundationmodel.DiscussCommentViewEdit,
	error,
) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":        1,
				"discuss_id": 1,
				"author_id":  1,
				"content":    1,
			},
		)
	var discussComment foundationmodel.DiscussCommentViewEdit
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&discussComment); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discuss comment error")
	}
	return &discussComment, nil
}

func (d *DiscussCommentDao) GetDiscussComment(ctx context.Context, id int) (*foundationmodel.DiscussComment, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":         1,
				"content":     1,
				"author_id":   1,
				"insert_time": 1,
				"update_time": 1,
			},
		)
	var discussComment foundationmodel.DiscussComment
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&discussComment); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discussComment error")
	}
	return &discussComment, nil
}

func (d *DiscussCommentDao) GetDiscussCommentList(
	ctx context.Context,
	discussId int,
	page int,
	pageSize int,
) (
	[]*foundationmodel.DiscussComment,
	int,
	error,
) {
	filter := bson.M{
		"discuss_id": discussId,
	}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(
			bson.M{
				"_id":         1,
				"content":     1,
				"author_id":   1,
				"insert_time": 1,
				"update_time": 1,
			},
		).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"_id": 1})
	// 查询总记录数
	totalCount, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count documents, page: %d", page)
	}
	// 查询当前页的数据
	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to find documents, page: %d", page)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	var list []*foundationmodel.DiscussComment
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *DiscussCommentDao) InsertDiscussComment(
	ctx context.Context,
	discussComment *foundationmodel.DiscussComment,
) error {
	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(
		ctx, func(sc mongo.SessionContext) (interface{}, error) {
			// 更新Discuss表的comment计数
			filter := bson.M{
				"_id": discussComment.DiscussId,
			}
			update := bson.M{
				"$set": bson.M{
					"update_time": discussComment.UpdateTime,
				},
			}
			res, err := GetDiscussDao().collection.UpdateOne(sc, filter, update)
			if err != nil {
				return nil, err
			}
			if res.MatchedCount == 0 {
				return nil, metaerror.New(
					"update discuss comment no document matched, discussId:%d",
					discussComment.DiscussId,
				)
			}
			// 获取下一个序列号
			seq, err := GetCounterDao().GetNextSequence(sc, "discuss_comment_id")
			if err != nil {
				return nil, err
			}
			// 更新 DiscussComment 的 ID
			discussComment.Id = seq
			// 插入新的 DiscussComment
			_, err = d.collection.InsertOne(sc, discussComment)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DiscussCommentDao) UpdateDiscussComments(ctx context.Context, tags []*foundationmodel.DiscussComment) error {
	var models []mongo.WriteModel
	for _, tab := range tags {
		filter := bson.D{
			{"_id", tab.Id},
		}
		update := bson.M{
			"$set": tab,
		}
		updateModel := mongo.NewUpdateManyModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		models = append(models, updateModel)
	}
	bulkOptions := options.BulkWrite().SetOrdered(false) // 设置是否按顺序执行
	_, err := d.collection.BulkWrite(ctx, models, bulkOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to perform bulk update")
	}
	return nil
}

func (d *DiscussCommentDao) UpdateContent(
	ctx *gin.Context,
	id int,
	discussId int,
	content string,
	updateTime time.Time,
) error {

	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(
		ctx, func(sc mongo.SessionContext) (interface{}, error) {
			filter := bson.M{
				"_id": discussId,
			}
			update := bson.M{
				"$set": bson.M{
					"update_time": updateTime,
				},
			}
			res, err := GetDiscussDao().collection.UpdateOne(sc, filter, update)
			if err != nil {
				return nil, err
			}
			if res.MatchedCount == 0 {
				return nil, metaerror.New("update discuss comment no document matched, discussId:%d", discussId)
			}
			// 更新 DiscussComment 的内容
			filter = bson.M{
				"_id": id,
			}
			update = bson.M{
				"$set": bson.M{
					"content":     content,
					"update_time": updateTime,
				},
			}
			res, err = d.collection.UpdateOne(sc, filter, update)
			if err != nil {
				return nil, metaerror.Wrap(err, "update discuss comment content failed, id: %d", id)
			}
			if res.MatchedCount == 0 {
				return nil, metaerror.New("update discuss comment no document matched, id:%d", id)
			}
			return nil, nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}
