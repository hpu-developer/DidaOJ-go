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
)

type JudgerDao struct {
	collection *mongo.Collection
}

var singletonJudgerDao = singleton.Singleton[JudgerDao]{}

func GetJudgerDao() *JudgerDao {
	return singletonJudgerDao.GetInstance(
		func() *JudgerDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var JudgerDao JudgerDao
			JudgerDao.collection = client.
				Database("didaoj").
				Collection("judger")
			return &JudgerDao
		},
	)
}

func (d *JudgerDao) InitDao(ctx context.Context) error {

	return nil
}

func (d *JudgerDao) GetJudgers(ctx context.Context) ([]*foundationmodel.Judger, error) {
	filter := bson.M{}
	opts := options.Find().
		SetSort(
			bson.D{
				{"_id", 1},
			},
		)
	// 查询当前页的数据
	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, metaerror.Wrap(err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	var list []*foundationmodel.Judger
	if err = cursor.All(ctx, &list); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode cursor")
	}
	return list, nil
}

func (d *JudgerDao) GetJudgerName(ctx context.Context, judger string) (*string, error) {
	filter := bson.D{
		{"_id", judger},
	}
	opts := options.FindOne().
		SetProjection(
			bson.D{
				{"name", 1},
			},
		)
	var result struct {
		Name string `bson:"name"`
	}
	err := d.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No judger found
		}
		return nil, metaerror.Wrap(err, "failed to find judger")
	}
	return &result.Name, nil
}

func (d *JudgerDao) UpdateJudger(ctx context.Context, judger *foundationmodel.Judger) error {
	filter := bson.D{
		{"_id", judger.Key},
	}
	update := bson.M{
		"$set": judger,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judger")
	}
	return nil
}
