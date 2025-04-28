package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	metamongo "meta/mongo"
	"meta/singleton"
)

type ProblemTagDao struct {
	collection *mongo.Collection
}

var singletonProblemTagDao = singleton.Singleton[ProblemTagDao]{}

func GetProblemTagDao() *ProblemTagDao {
	return singletonProblemTagDao.GetInstance(
		func() *ProblemTagDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var ProblemTagDao ProblemTagDao
			ProblemTagDao.collection = client.
				Database("didaoj").
				Collection("problem_tag")
			return &ProblemTagDao
		},
	)
}

func (d *ProblemTagDao) InitDao(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}}, // 1表示升序索引
		Options: options.Index().SetUnique(true),
	}
	_, err := d.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProblemTagDao) UpdateProblemTag(ctx context.Context, key string, problemTag *foundationmodel.ProblemTag) error {
	filter := bson.D{
		{"_id", key},
	}
	update := bson.M{
		"$set": problemTag,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *ProblemTagDao) GetProblemTag(ctx context.Context, key string) (*foundationmodel.ProblemTag, error) {
	filter := bson.M{
		"_id": key,
	}
	var problemTag foundationmodel.ProblemTag
	if err := d.collection.FindOne(ctx, filter).Decode(&problemTag); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problemTag error")
	}
	return &problemTag, nil
}

func (d *ProblemTagDao) GetProblemTagList(ctx context.Context) ([]foundationmodel.ProblemTag, error) {
	filter := bson.M{}
	cursor, err := d.collection.Find(ctx, filter, options.Find())
	if err != nil {
		return nil, metaerror.Wrap(err, "find ProblemTag error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(cursor, ctx)
	var problemList []foundationmodel.ProblemTag
	if err = cursor.All(ctx, &problemList); err != nil {
		return nil, metaerror.Wrap(err, "decode ProblemTag error")
	}
	return problemList, nil
}

func (d *ProblemTagDao) UpdateProblemTags(ctx context.Context, tags []*foundationmodel.ProblemTag) error {
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
