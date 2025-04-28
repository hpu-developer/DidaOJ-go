package foundationdao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metamongo "meta/meta-mongo"
	"meta/singleton"
)

type CounterDao struct {
	collection *mongo.Collection
}

var singletonCounterDao = singleton.Singleton[CounterDao]{}

func GetCounterDao() *CounterDao {
	return singletonCounterDao.GetInstance(
		func() *CounterDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var CounterDao CounterDao
			CounterDao.collection = client.
				Database("didaoj").
				Collection("counter")
			return &CounterDao
		},
	)
}

func (d *CounterDao) InitDao(ctx context.Context) error {
	err := d.collection.Drop(ctx)
	if err != nil {
		return err
	}
	err = d.InitCounter(ctx, "problem_id", 999)
	if err != nil {
		return err
	}
	err = d.InitCounter(ctx, "status", 0)
	if err != nil {
		return err
	}
	return nil
}

func (d *CounterDao) InitCounter(ctx context.Context, key string, seq int) error {
	_, err := d.collection.InsertOne(ctx, bson.M{"_id": key, "seq": seq})
	if err != nil {
		return err
	}
	return nil
}

func (d *CounterDao) GetNextSequence(ctx context.Context, key string) (int, error) {
	filter := bson.M{"_id": key}
	update := bson.M{
		"$inc": bson.M{"seq": 1},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var result struct {
		Seq int `bson:"seq"`
	}
	err := d.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}
