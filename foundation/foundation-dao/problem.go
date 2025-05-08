package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
)

type ProblemDao struct {
	collection *mongo.Collection
}

var singletonProblemDao = singleton.Singleton[ProblemDao]{}

func GetProblemDao() *ProblemDao {
	return singletonProblemDao.GetInstance(
		func() *ProblemDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var ProblemDao ProblemDao
			ProblemDao.collection = client.
				Database("didaoj").
				Collection("problem")
			return &ProblemDao
		},
	)
}

func (d *ProblemDao) InitDao(ctx context.Context) error {
	return nil
}

func (d *ProblemDao) UpdateProblem(ctx context.Context, key string, problem *foundationmodel.Problem) error {
	filter := bson.D{
		{"_id", key},
	}
	update := bson.M{
		"$set": problem,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *ProblemDao) GetProblem(ctx context.Context, key string) (*foundationmodel.Problem, error) {
	filter := bson.M{
		"_id": key,
	}
	var problem foundationmodel.Problem
	if err := d.collection.FindOne(ctx, filter).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemList(ctx context.Context,
	page int,
	pageSize int,
) ([]*foundationmodel.Problem,
	int,
	error,
) {
	filter := bson.M{}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(bson.M{
			"_id":     1,
			"title":   1,
			"tags":    1,
			"accept":  1,
			"attempt": 1,
		}).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{
			{Key: "sort", Value: 1},
			{Key: "_id", Value: 1},
		})
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
	var list []*foundationmodel.Problem
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *ProblemDao) UpdateProblems(ctx context.Context, tags []*foundationmodel.Problem) error {
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
