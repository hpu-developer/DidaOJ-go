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

func (d *ProblemDao) GetProblemList(ctx context.Context) ([]foundationmodel.Problem, error) {
	filter := bson.M{}
	cursor, err := d.collection.Find(ctx, filter, options.Find())
	if err != nil {
		return nil, metaerror.Wrap(err, "find Problem error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(cursor, ctx)
	var problemList []foundationmodel.Problem
	if err = cursor.All(ctx, &problemList); err != nil {
		return nil, metaerror.Wrap(err, "decode Problem error")
	}
	return problemList, nil
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
