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
	return nil
}

func (d *ProblemDailyDao) GetProblemIdByDaily(ctx *gin.Context, id string) (*string, error) {
	filter := bson.D{
		{"_id", id},
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

func (d *ProblemDailyDao) GetDailyList(
	ctx *gin.Context,
	startDate *string,
	endDate *string,
	page int,
	pageSize int,
) ([]*foundationmodel.ProblemDaily, int, error) {
	filter := bson.M{}
	if startDate != nil && *startDate != "" {
		filter["_id"] = bson.M{
			"$gte": *startDate,
		}
	}
	if endDate != nil && *endDate != "" {
		if _, ok := filter["_id"]; !ok {
			filter["_id"] = bson.M{}
		}
		filter["_id"].(bson.M)["$lte"] = *endDate
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
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(7)
	cursor, err := d.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to find problem daily")
	}
	var problemDailies []*foundationmodel.ProblemDaily
	if err := cursor.All(ctx, &problemDailies); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode problem daily")
	}
	return problemDailies, nil
}

func (d *ProblemDailyDao) UpdateProblemDaily(
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
