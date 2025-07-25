package foundationdaomongo

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model-mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"regexp"
)

type DiscussTagDao struct {
	collection *mongo.Collection
}

var singletonDiscussTagDao = singleton.Singleton[DiscussTagDao]{}

func GetDiscussTagDao() *DiscussTagDao {
	return singletonDiscussTagDao.GetInstance(
		func() *DiscussTagDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var DiscussTagDao DiscussTagDao
			DiscussTagDao.collection = client.
				Database("didaoj").
				Collection("discuss_tag")
			return &DiscussTagDao
		},
	)
}

func (d *DiscussTagDao) InitDao(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}}, // 1表示升序索引
		Options: options.Index().SetUnique(true),
	}
	_, err := d.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return nil
}

func (d *DiscussTagDao) UpdateDiscussTag(
	ctx context.Context,
	key string,
	discussTag *foundationmodel.DiscussTag,
) error {
	filter := bson.D{
		{"_id", key},
	}
	update := bson.M{
		"$set": discussTag,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *DiscussTagDao) GetDiscussTag(ctx context.Context, key string) (*foundationmodel.DiscussTag, error) {
	filter := bson.M{
		"_id": key,
	}
	var discussTag foundationmodel.DiscussTag
	if err := d.collection.FindOne(ctx, filter).Decode(&discussTag); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discussTag error")
	}
	return &discussTag, nil
}

func (d *DiscussTagDao) GetDiscussTagList(ctx context.Context, maxCount int) (
	[]*foundationmodel.DiscussTag,
	int,
	error,
) {
	filter := bson.M{}
	findOptions := options.Find().
		SetProjection(
			bson.M{
				"update_time": 0,
			},
		).
		SetSort(bson.D{{Key: "update_time", Value: -1}})
	if maxCount > 0 {
		findOptions.SetLimit(int64(maxCount))
	}
	totalCount, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count documents, maxCount: %d", maxCount)
	}
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "find DiscussTag error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(cursor, ctx)
	var discussList []*foundationmodel.DiscussTag
	if err = cursor.All(ctx, &discussList); err != nil {
		return nil, 0, metaerror.Wrap(err, "decode DiscussTag error")
	}
	return discussList, int(totalCount), nil
}

func (d *DiscussTagDao) UpdateDiscussTags(ctx context.Context, tags []*foundationmodel.DiscussTag) error {
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

func (d *DiscussTagDao) GetDiscussTagByIds(ctx context.Context, ids []int) ([]*foundationmodel.DiscussTag, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	findOptions := options.Find().
		SetProjection(
			bson.M{
				"update_time": 0,
			},
		).
		SetSort(bson.D{{Key: "update_time", Value: -1}})
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, metaerror.Wrap(err, "find DiscussTag error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(cursor, ctx)
	var tagList []*foundationmodel.DiscussTag
	if err = cursor.All(ctx, &tagList); err != nil {
		return nil, metaerror.Wrap(err, "decode DiscussTag error")
	}
	return tagList, nil
}

func (d *DiscussTagDao) SearchTags(ctx context.Context, tag string) ([]int, error) {
	filter := bson.M{
		"name": bson.M{
			"$regex":   regexp.QuoteMeta(tag),
			"$options": "i", // 不区分大小写
		},
	}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, metaerror.Wrap(err, "find user account info error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err, "close cursor error")
		}
	}(cursor, ctx)
	var result []int
	for cursor.Next(ctx) {
		var discusss foundationmodel.DiscussTag
		if err := cursor.Decode(&discusss); err != nil {
			return nil, metaerror.Wrap(err, "decode user account info error")
		}
		result = append(result, discusss.Id)
	}
	return result, nil
}

func (d *DiscussTagDao) InsertTag(ctx context.Context, tag *foundationmodel.DiscussTag) error {
	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(
		ctx, func(sc mongo.SessionContext) (interface{}, error) {
			seq, err := GetCounterDao().GetNextSequence(sc, "discuss_tag_id")
			if err != nil {
				return nil, err
			}
			tag.Id = int(seq)
			_, err = d.collection.InsertOne(sc, tag)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
	if mongo.IsDuplicateKeyError(err) {
		err := d.collection.FindOne(ctx, bson.M{"name": tag.Name}).Decode(tag)
		if err != nil {
			return metaerror.Wrap(err, "find tag error, name:%s", tag.Name)
		}
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
