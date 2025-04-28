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

type JudgeJobDao struct {
	collection *mongo.Collection
}

var singletonJudgeJobDao = singleton.Singleton[JudgeJobDao]{}

func GetJudgeJobDao() *JudgeJobDao {
	return singletonJudgeJobDao.GetInstance(
		func() *JudgeJobDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var JudgeJobDao JudgeJobDao
			JudgeJobDao.collection = client.
				Database("didaoj").
				Collection("judge_job")
			return &JudgeJobDao
		},
	)
}

func (d *JudgeJobDao) InitDao(ctx context.Context) error {
	return nil
}

func (d *JudgeJobDao) UpdateJudgeJob(ctx context.Context, id int, judgeSource *foundationmodel.JudgeJob) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": judgeSource,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *JudgeJobDao) GetJudgeJob(ctx context.Context, key string) (*foundationmodel.JudgeJob, error) {
	filter := bson.M{
		"_id": key,
	}
	var judgeSource foundationmodel.JudgeJob
	if err := d.collection.FindOne(ctx, filter).Decode(&judgeSource); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find judgeSource error")
	}
	return &judgeSource, nil
}

func (d *JudgeJobDao) GetJudgeJobList(ctx context.Context) ([]foundationmodel.JudgeJob, error) {
	filter := bson.M{}
	cursor, err := d.collection.Find(ctx, filter, options.Find())
	if err != nil {
		return nil, metaerror.Wrap(err, "find JudgeJob error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(cursor, ctx)
	var judgeSourceList []foundationmodel.JudgeJob
	if err = cursor.All(ctx, &judgeSourceList); err != nil {
		return nil, metaerror.Wrap(err, "decode JudgeJob error")
	}
	return judgeSourceList, nil
}

func (d *JudgeJobDao) UpdateJudgeJobs(ctx context.Context, tags []*foundationmodel.JudgeJob) error {
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
