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
)

type CollectionDao struct {
	collection *mongo.Collection
}

var singletonCollectionDao = singleton.Singleton[CollectionDao]{}

func GetCollectionDao() *CollectionDao {
	return singletonCollectionDao.GetInstance(
		func() *CollectionDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var CollectionDao CollectionDao
			CollectionDao.collection = client.
				Database("didaoj").
				Collection("collection")
			return &CollectionDao
		},
	)
}

func (d *CollectionDao) InitDao(ctx context.Context) error {
	return nil
}

func (d *CollectionDao) HasCollectionTitle(ctx *gin.Context, ownerId int, title string) (bool, error) {
	filter := bson.M{
		"title":    title,
		"owner_id": bson.M{"$ne": ownerId},
	}
	count, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, metaerror.Wrap(err, "failed to count documents")
	}
	return count > 0, nil
}

func (d *CollectionDao) CheckJoinAuth(ctx *gin.Context, collectionId int, userId int) (bool, error) {
	filter := bson.D{
		{"_id", collectionId},
		{
			"$or", bson.A{
				bson.D{{"owner_id", userId}},
				bson.D{{"private", bson.M{"$exists": false}}},
			},
		},
	}
	count, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *CollectionDao) GetListAll(ctx context.Context) ([]*foundationmodel.Collection, error) {
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := d.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find all collections error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var collections []*foundationmodel.Collection
	for cursor.Next(ctx) {
		var collection foundationmodel.Collection
		if err := cursor.Decode(&collection); err != nil {
			return nil, metaerror.Wrap(err, "decode collection error")
		}
		collections = append(collections, &collection)
	}
	return collections, nil
}

func (d *CollectionDao) GetCollection(ctx context.Context, id int) (*foundationmodel.Collection, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":         1,
				"title":       1,
				"description": 1,
				"start_time":  1,
				"end_time":    1,
				"owner_id":    1,
				"create_time": 1,
				"update_time": 1,
				"problems":    1,
				"private":     1,
				"members":     1,
			},
		)
	var collection foundationmodel.Collection
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&collection); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find collection error")
	}
	return &collection, nil
}

func (d *CollectionDao) GetCollectionEdit(ctx context.Context, id int) (*foundationmodel.Collection, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":         1,
				"title":       1,
				"description": 1,
				"start_time":  1,
				"end_time":    1,
				"owner_id":    1,
				"create_time": 1,
				"update_time": 1,
				"problems":    1,
				"private":     1,
				"members":     1,
			},
		)
	var collection foundationmodel.Collection
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&collection); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find collection error")
	}
	return &collection, nil
}

func (d *CollectionDao) GetCollectionOwnerId(ctx context.Context, id int) (int, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":      1,
				"owner_id": 1,
			},
		)
	var collection struct {
		Id      int `bson:"_id"`
		OwnerId int `bson:"owner_id"`
	}
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&collection); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, metaerror.Wrap(err, "find collection error")
	}
	return collection.OwnerId, nil
}

func (d *CollectionDao) GetCollectionTitle(ctx context.Context, id int) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":   1,
				"title": 1,
			},
		)
	var collection foundationmodel.Collection
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&collection); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find collection error")
	}
	return &collection.Title, nil
}

func (d *CollectionDao) GetProblems(ctx context.Context, id int) ([]string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"problems": 1,
		},
	)
	var result struct {
		Problems []string `bson:"problems"`
	}
	err := d.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problems error")
	}
	return result.Problems, nil
}

func (d *CollectionDao) GetCollectionList(
	ctx context.Context,
	page int,
	pageSize int,
) (
	[]*foundationmodel.Collection,
	int,
	error,
) {
	filter := bson.M{}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(
			bson.M{
				"_id":        1,
				"title":      1,
				"start_time": 1,
				"end_time":   1,
				"owner_id":   1,
			},
		).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"_id": -1})
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
	var list []*foundationmodel.Collection
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *CollectionDao) GetCollectionRankView(ctx context.Context, id int) (
	*foundationmodel.CollectionRankView,
	error,
) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":        1,
				"start_time": 1,
				"end_time":   1,
				"problems":   1,
				"members":    1,
			},
		)
	var collection foundationmodel.CollectionRankView
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&collection); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find collection error")
	}
	return &collection, nil
}

func (d *CollectionDao) PostJoin(ctx *gin.Context, collectionId int, userId int) error {
	filter := bson.D{
		{"_id", collectionId},
	}
	update := bson.M{
		"$addToSet": bson.M{"members": userId},
	}
	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update collection members")
	}
	if result.MatchedCount == 0 {
		return metaerror.New("collection not found")
	}
	return nil
}

func (d *CollectionDao) PostQuit(ctx *gin.Context, collectionId int, userId int) error {
	filter := bson.D{
		{"_id", collectionId},
	}
	update := bson.M{
		"$pull": bson.M{"members": userId},
	}
	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update collection members")
	}
	if result.MatchedCount == 0 {
		return metaerror.New("collection not found")
	}
	return nil
}

func (d *CollectionDao) UpdateCollection(
	ctx context.Context,
	id int,
	collection *foundationmodel.Collection,
) error {
	filter := bson.D{
		{"_id", id},
	}
	setData := metamongo.StructToMapInclude(
		collection,
		"title",
		"description",
		"start_time",
		"end_time",
		"problems",
		"private",
		"members",
		"update_time",
	)
	update := bson.M{
		"$set": setData,
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *CollectionDao) UpdateCollections(ctx context.Context, tags []*foundationmodel.Collection) error {
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

func (d *CollectionDao) InsertCollection(ctx context.Context, collection *foundationmodel.Collection) error {
	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(
		ctx, func(sc mongo.SessionContext) (interface{}, error) {
			// 获取下一个序列号
			seq, err := GetCounterDao().GetNextSequence(sc, "collection_id")
			if err != nil {
				return nil, err
			}
			// 更新 Collection 的 ID
			collection.Id = seq
			// 插入新的 Collection
			_, err = d.collection.InsertOne(sc, collection)
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
