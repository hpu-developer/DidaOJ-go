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
	"regexp"
)

type DiscussDao struct {
	collection *mongo.Collection
}

var singletonDiscussDao = singleton.Singleton[DiscussDao]{}

func GetDiscussDao() *DiscussDao {
	return singletonDiscussDao.GetInstance(
		func() *DiscussDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var DiscussDao DiscussDao
			DiscussDao.collection = client.
				Database("didaoj").
				Collection("discuss")
			return &DiscussDao
		},
	)
}

func (d *DiscussDao) InitDao(ctx context.Context) error {
	return nil
}

func (d *DiscussDao) GetListAll(ctx context.Context) ([]*foundationmodel.Discuss, error) {
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := d.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find all contests error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var contests []*foundationmodel.Discuss
	for cursor.Next(ctx) {
		var contest foundationmodel.Discuss
		if err := cursor.Decode(&contest); err != nil {
			return nil, metaerror.Wrap(err, "decode contest error")
		}
		contests = append(contests, &contest)
	}
	return contests, nil
}

func (d *DiscussDao) GetAuthorId(ctx context.Context, id int) (int, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"author_id": 1,
			},
		)
	var discuss foundationmodel.Discuss
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&discuss); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, metaerror.Wrap(err, "find discuss error")
	}
	return discuss.AuthorId, nil
}

func (d *DiscussDao) IsDiscussBannedOrNotExist(ctx context.Context, id int) (bool, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"banned": 1,
			},
		)
	var discuss struct {
		Banned bool `bson:"banned"`
	}
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&discuss); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return true, nil
		}
		return false, metaerror.Wrap(err, "find discuss error")
	}
	return discuss.Banned, nil
}

func (d *DiscussDao) GetContent(ctx context.Context, id int) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"content": 1,
			},
		)
	var discuss foundationmodel.Discuss
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&discuss); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discuss content error")
	}
	return &discuss.Content, nil
}

func (d *DiscussDao) GetDiscuss(ctx context.Context, id int) (*foundationmodel.Discuss, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":         1,
				"title":       1,
				"content":     1,
				"insert_time": 1,
				"modify_time": 1,
				"update_time": 1,
				"author_id":   1,
				"create_time": 1,
				"view_count":  1,
				"tags":        1,
				"problem_id":  1,
				"contest_id":  1,
				"judge_id":    1,
			},
		)
	var discuss foundationmodel.Discuss
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&discuss); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find discuss error")
	}
	return &discuss, nil
}

func (d *DiscussDao) GetDiscussList(
	ctx context.Context,
	onlyProblem bool,
	contestId int,
	problemId string,
	title string,
	userId int,
	page int,
	pageSize int,
) ([]*foundationmodel.Discuss, int, error) {
	filter := bson.M{}
	if contestId > 0 {
		filter["contest_id"] = contestId
	} else {
		filter["contest_id"] = bson.M{"$exists": false}
	}
	if problemId != "" {
		filter["problem_id"] = problemId
	} else {
		if onlyProblem {
			filter["problem_id"] = bson.M{"$exists": true}
		}
	}
	if title != "" {
		filter["title"] = bson.M{
			"$regex":   regexp.QuoteMeta(title),
			"$options": "i", // 不区分大小写
		}
	}
	if userId > 0 {
		filter["author_id"] = userId
	}

	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	projection := bson.M{
		"_id":         1,
		"title":       1,
		"insert_time": 1,
		"modify_time": 1,
		"update_time": 1,
		"author_id":   1,
		"view_count":  1,

		"contest_id": 1,
		"problem_id": 1,
	}

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(projection).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"update_time": -1})
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
	var list []*foundationmodel.Discuss
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *DiscussDao) UpdateContent(ctx *gin.Context, id int, description string) error {
	filter := bson.M{
		"_id": id,
	}
	update := bson.M{
		"$set": bson.M{
			"content": description,
		},
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update discuss content")
	}
	return nil
}

func (d *DiscussDao) UpdateDiscuss(ctx context.Context, discussId int, discuss *foundationmodel.Discuss) error {
	filter := bson.D{
		{"_id", discussId},
	}
	setData := metamongo.StructToMapInclude(
		discuss,
		"title",
		"content",
		"modify_time",
		"update_time",
	)
	unsetData := bson.M{}
	if discuss.ProblemId != nil {
		setData["problem_id"] = discuss.ProblemId
	} else {
		unsetData["problem_id"] = 1
	}
	update := bson.M{
		"$set":   setData,
		"$unset": unsetData,
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to save discuss")
	}
	return nil
}

func (d *DiscussDao) UpdateDiscusses(ctx context.Context, tags []*foundationmodel.Discuss) error {
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

func (d *DiscussDao) InsertDiscuss(ctx context.Context, discuss *foundationmodel.Discuss) error {
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
			seq, err := GetCounterDao().GetNextSequence(sc, "discuss_id")
			if err != nil {
				return nil, err
			}
			// 更新 Discuss 的 ID
			discuss.Id = seq
			// 插入新的 Discuss
			_, err = d.collection.InsertOne(sc, discuss)
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
