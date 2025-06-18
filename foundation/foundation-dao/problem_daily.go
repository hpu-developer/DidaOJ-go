package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
)

type ProblemDailyDao struct {
	collection *mongo.Collection
}

var singletonProblemDailyDao = singleton.Singleton[ProblemDailyDao]{}

func GetProblemDailyDao() *ProblemDailyDao {
	return singletonProblemDailyDao.GetInstance(
		func() *ProblemDailyDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var ProblemDailyDao ProblemDailyDao
			ProblemDailyDao.collection = client.
				Database("didaoj").
				Collection("problem_daily")
			return &ProblemDailyDao
		},
	)
}

func (d *ProblemDailyDao) InitDao(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "problem_id", Value: 1}}, // 1表示升序索引
		Options: options.Index().SetUnique(true),
	}
	_, err := d.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProblemDailyDao) GetProblemIdByDaily(ctx *gin.Context, id string) (*string, error) {
	nowId := metatime.GetTimeNow().Format("2006-01-02")
	filter := bson.M{
		"_id": bson.M{
			"$lte": nowId,
			"$eq":  id,
		},
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"problem_id": 1,
			},
		)
	var result struct {
		ProblemId *string `bson:"problem_id"`
	}
	err := d.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 没有找到对应的记录
		}
		return nil, metaerror.Wrap(err, "failed to find problem id by daily id: %s", id)
	}
	return result.ProblemId, nil
}

func (d *ProblemDailyDao) GetProblemDaily(ctx *gin.Context, id string) (*foundationmodel.ProblemDaily, error) {
	nowId := metatime.GetTimeNow().Format("2006-01-02")
	filter := bson.M{
		"_id": bson.M{
			"$lte": nowId,
			"$eq":  id,
		},
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":        1,
				"problem_id": 1,
				"solution":   1,
				"code":       1,
			},
		)
	var problemDaily foundationmodel.ProblemDaily
	err := d.collection.FindOne(ctx, filter, opts).Decode(&problemDaily)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 没有找到对应的记录
		}
		return nil, metaerror.Wrap(err, "failed to find problem daily by id: %s", id)
	}
	return &problemDaily, nil
}

func (d *ProblemDailyDao) GetProblemDailyEdit(ctx *gin.Context, id string) (*foundationmodel.ProblemDaily, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":         1,
				"problem_id":  1,
				"solution":    1,
				"code":        1,
				"create_time": 1,
				"update_time": 1,
				"creator_id":  1,
				"updater_id":  1,
			},
		)
	var problemDaily foundationmodel.ProblemDaily
	err := d.collection.FindOne(ctx, filter, opts).Decode(&problemDaily)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 没有找到对应的记录
		}
		return nil, metaerror.Wrap(err, "failed to find problem daily by id: %s", id)
	}
	return &problemDaily, nil
}

func (d *ProblemDailyDao) GetDailyList(
	ctx *gin.Context,
	startDate *string,
	endDate *string,
	page int,
	pageSize int,
) ([]*foundationmodel.ProblemDaily, int, error) {
	nowId := metatime.GetTimeNow().Format("2006-01-02")
	idFilter := bson.M{}
	if startDate != nil && *startDate != "" {
		idFilter["$gte"] = *startDate
	}
	if endDate != nil && *endDate != "" {
		if *endDate < nowId {
			idFilter["$lte"] = *endDate
		} else {
			idFilter["$lte"] = nowId
		}
	} else {
		idFilter["$lte"] = nowId
	}
	filter := bson.M{
		"_id": idFilter,
	}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(
			bson.M{
				"_id":        1,
				"problem_id": 1,
			},
		).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(
			bson.D{
				{"_id", -1},
			},
		)
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
	var list []*foundationmodel.ProblemDaily
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *ProblemDailyDao) GetDailyRecently(ctx *gin.Context) ([]*foundationmodel.ProblemDaily, error) {
	filter := bson.M{
		"_id": bson.M{
			"$lte": metatime.GetTimeNow().Format("2006-01-02"),
		},
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(7)
	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to find problem daily")
	}
	var problemDailies []*foundationmodel.ProblemDaily
	if err := cursor.All(ctx, &problemDailies); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode problem daily")
	}
	return problemDailies, nil
}

func (d *ProblemDailyDao) PostDailyCreate(ctx *gin.Context, problemDaily *foundationmodel.ProblemDaily) error {
	_, err := d.collection.InsertOne(ctx, problemDaily)
	if err != nil {
		return metaerror.Wrap(err, "failed to insert problem daily")
	}
	return nil
}

func (d *ProblemDailyDao) UpdateProblemDaily(
	ctx context.Context,
	id string,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	filter := bson.D{
		{"_id", id},
	}
	setData := metamongo.StructToMapInclude(
		problemDaily,
		"problem_id",
		"solution",
		"code",
		"update_time",
		"updater_id",
	)
	update := bson.M{
		"$set": setData,
	}
	res, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	if res.MatchedCount == 0 {
		return metaerror.New("no document matched for update, id: %s", id)
	}
	return nil
}

func (d *ProblemDailyDao) UpdateOrInsertProblemDaily(
	ctx context.Context,
	id string,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": problemDaily,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}
