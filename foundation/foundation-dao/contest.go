package foundationdao

import (
	"context"
	"errors"
	"fmt"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
)

type ContestDao struct {
	collection *mongo.Collection
}

var singletonContestDao = singleton.Singleton[ContestDao]{}

func GetContestDao() *ContestDao {
	return singletonContestDao.GetInstance(
		func() *ContestDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var ContestDao ContestDao
			ContestDao.collection = client.
				Database("didaoj").
				Collection("contest")
			return &ContestDao
		},
	)
}

func (d *ContestDao) InitDao(ctx context.Context) error {
	return nil
}

func (d *ContestDao) UpdateContest(ctx context.Context, key string, contest *foundationmodel.Contest) error {
	filter := bson.D{
		{"_id", key},
	}
	update := bson.M{
		"$set": contest,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *ContestDao) GetContest(ctx context.Context, id int) (*foundationmodel.Contest, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(bson.M{
			"_id":          1,
			"title":        1,
			"start_time":   1,
			"end_time":     1,
			"owner_id":     1,
			"create_time":  1,
			"problems":     1,
			"auth":         1,
			"type":         1,
			"score_type":   1,
			"always_lock":  1,
			"descriptions": 1,
			"notification": 1,
		})
	var contest foundationmodel.Contest
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest, nil
}

func (d *ContestDao) GetContestTitle(ctx context.Context, id int) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(bson.M{
			"_id":   1,
			"title": 1,
		})
	var contest foundationmodel.Contest
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest.Title, nil
}

func (d *ContestDao) GetContestRankView(ctx context.Context, id int) (*foundationmodel.ContestRankView, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(bson.M{
			"_id":        1,
			"start_time": 1,
			"end_time":   1,
			"problems":   1,
		})
	var contest foundationmodel.Contest
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &foundationmodel.ContestRankView{
		Id:        contest.Id,
		StartTime: &contest.StartTime,
		EndTime:   &contest.EndTime,
		Problems:  contest.Problems,
	}, nil
}

func (d *ContestDao) GetProblems(ctx context.Context, id int) ([]*foundationmodel.ContestProblem, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(bson.M{
		"problems": 1,
	})
	var result struct {
		Problems []*foundationmodel.ContestProblem `bson:"problems"`
	}
	err := d.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problems error")
	}
	return result.Problems, nil
}

func (d *ContestDao) GetProblemIndex(ctx context.Context, id int, problemId *string) (int, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(bson.M{
		"problems": bson.M{
			"$elemMatch": bson.M{
				"problem_id": *problemId,
			},
		},
	})
	var result struct {
		Problems []struct {
			Index int `bson:"index"`
		} `bson:"problems"`
	}
	err := d.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	if len(result.Problems) == 0 {
		return 0, fmt.Errorf("problem_id %s not found", *problemId)
	}
	return result.Problems[0].Index, nil
}

func (d *ContestDao) GetProblemIdByContest(ctx context.Context, id int, problemIndex int) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(bson.M{
		"problems": bson.M{
			"$elemMatch": bson.M{
				"index": problemIndex,
			},
		},
	})
	var result struct {
		Problems []struct {
			ProblemId string `bson:"problem_id"`
		} `bson:"problems"`
	}
	err := d.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	if len(result.Problems) == 0 {
		return nil, fmt.Errorf("problem index %d not found", problemIndex)
	}
	return &result.Problems[0].ProblemId, nil
}

func (d *ContestDao) GetContestList(ctx context.Context,
	page int,
	pageSize int,
) ([]*foundationmodel.Contest,
	int,
	error,
) {
	filter := bson.M{}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(bson.M{
			"_id":        1,
			"title":      1,
			"start_time": 1,
			"end_time":   1,
			"owner_id":   1,
		}).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"_id": -1})
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
	var list []*foundationmodel.Contest
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *ContestDao) UpdateContests(ctx context.Context, tags []*foundationmodel.Contest) error {
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

func (d *ContestDao) InsertContest(ctx context.Context, contest *foundationmodel.Contest) error {
	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		// 获取下一个序列号
		seq, err := GetCounterDao().GetNextSequence(sc, "contest_id")
		if err != nil {
			return nil, err
		}
		// 更新 Contest 的 ID
		contest.Id = seq
		// 插入新的 Contest
		_, err = d.collection.InsertOne(sc, contest)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
