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
	metatime "meta/meta-time"
	metatype "meta/meta-type"
	"meta/singleton"
	"web/request"
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
	_, err := d.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "sort", Value: 1},
				{Key: "_id", Value: 1},
			},
			Options: options.Index().SetName("idx_sort_id"),
		},
		{
			Keys: bson.D{
				{Key: "origin_oj", Value: 1},
				{Key: "origin_id", Value: 1},
			},
			Options: options.Index().SetName("idx_origin_id"),
		},
	})
	if err != nil {
		return metaerror.Wrap(err, "failed to create index for problem collection")
	}
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

func (d *ProblemDao) HasProblem(ctx context.Context, id string) (bool, error) {
	filter := bson.M{
		"_id": id,
	}
	count, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, metaerror.Wrap(err, "failed to count documents")
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (d *ProblemDao) GetProblem(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	filter := bson.M{
		"_id": id,
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

func (d *ProblemDao) GetProblemJudge(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(bson.M{
		"_id":              1,
		"title":            1,
		"insert_time":      1,
		"update_time":      1,
		"creator_id":       1,
		"creator_nickname": 1,
		"judge_md5":        1,
	})
	var problem foundationmodel.Problem
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemListTitle(ctx context.Context, ids []string) ([]*foundationmodel.ProblemViewTitle, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1, "title": 1})
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, metaerror.Wrap(err, "find user account info error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err, "close cursor error")
		}
	}(cursor, ctx)
	var result []*foundationmodel.ProblemViewTitle
	for cursor.Next(ctx) {
		var problems foundationmodel.ProblemViewTitle
		if err := cursor.Decode(&problems); err != nil {
			return nil, metaerror.Wrap(err, "decode user account info error")
		}
		result = append(result, &problems)
	}
	return result, nil
}

func (d *ProblemDao) GetProblemList(ctx context.Context,
	title string, tags []int,
	page int,
	pageSize int,
) ([]*foundationmodel.Problem,
	int,
	error,
) {
	filter := bson.M{}
	if title != "" {
		filter["title"] = bson.M{
			"$regex":   title,
			"$options": "i", // 不区分大小写
		}
	}
	if len(tags) > 0 {
		filter["tags"] = bson.M{
			"$in": tags,
		}
	}
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

func (d *ProblemDao) UpdateProblemsExcludeJudgeMd5(ctx context.Context, problems []*foundationmodel.Problem) error {
	var models []mongo.WriteModel
	for _, problem := range problems {
		setFields := metatype.StructToMapExclude(problem, "judge_md5")
		filter := bson.D{
			{"_id", problem.Id},
		}
		update := bson.M{
			"$set":         setFields,
			"$setOnInsert": bson.M{"judge_md5": problem.JudgeMd5},
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

func (d *ProblemDao) PostEdit(ctx context.Context, id int, data *request.ProblemEdit) error {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		var tagIds []int
		for _, tagName := range data.Tags {
			tag := foundationmodel.NewProblemTagBuilder().Name(tagName).Build()
			err := GetProblemTagDao().InsertTag(sessCtx, tag)
			if err != nil {
				return nil, err
			}
			tagIds = append(tagIds, tag.Id)
		}
		nowTime := metatime.GetTimeNow()
		_, err = d.collection.UpdateOne(sessCtx, bson.M{"_id": data.Id}, bson.M{
			"$set": bson.M{
				"title":        data.Title,
				"description":  data.Description,
				"tags":         tagIds,
				"update_time":  nowTime,
				"time_limit":   data.TimeLimit,
				"memory_limit": data.MemoryLimit,
				"source":       data.Source,
			},
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		return metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return nil
}

func (d *ProblemDao) UpdateJudgeMd5(ctx context.Context, id string, md5 string) error {
	filter := bson.M{
		"_id": id,
	}
	update := bson.M{
		"$set": bson.M{
			"judge_md5": md5,
		},
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judge md5")
	}
	return nil
}
