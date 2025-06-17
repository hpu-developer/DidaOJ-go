package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
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
