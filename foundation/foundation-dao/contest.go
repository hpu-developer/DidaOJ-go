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

func (d *ContestDao) HasContestTitle(ctx context.Context, ownerId int, title string) (bool, error) {
	filter := bson.M{
		"title":    title,
		"owner_id": bson.M{"$ne": ownerId},
	}
	count, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, metaerror.Wrap(err, "failed to count documents")
	}
	return count > 0, nil
}

func (d *ContestDao) GetContest(ctx context.Context, id int) (*foundationmodel.Contest, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":                1,
				"title":              1,
				"start_time":         1,
				"end_time":           1,
				"owner_id":           1,
				"create_time":        1,
				"update_time":        1,
				"description":        1,
				"problems":           1,
				"private":            1,
				"type":               1,
				"score_type":         1,
				"descriptions":       1,
				"notification":       1,
				"lock_rank_duration": 1,
				"always_lock":        1,
			},
		)
	var contest foundationmodel.Contest
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest, nil
}

func (d *ContestDao) GetContestEdit(ctx context.Context, id int) (*foundationmodel.Contest, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":                1,
				"title":              1,
				"start_time":         1,
				"end_time":           1,
				"owner_id":           1,
				"create_time":        1,
				"update_time":        1,
				"description":        1,
				"problems":           1,
				"private":            1,
				"type":               1,
				"score_type":         1,
				"descriptions":       1,
				"notification":       1,
				"members":            1,
				"lock_rank_duration": 1,
				"always_lock":        1,
			},
		)
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
		SetProjection(
			bson.M{
				"_id":   1,
				"title": 1,
			},
		)
	var contest foundationmodel.Contest
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest.Title, nil
}

func (d *ContestDao) GetContestViewLock(ctx context.Context, id int) (*foundationmodel.ContestViewLock, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":                1,
				"owner_id":           1,
				"auth_members":       1,
				"start_time":         1,
				"end_time":           1,
				"type":               1,
				"always_lock":        1,
				"lock_rank_duration": 1,
			},
		)
	var contest foundationmodel.ContestViewLock
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest, nil
}

func (d *ContestDao) GetContestViewRank(ctx context.Context, id int) (*foundationmodel.ContestViewRank, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":        1,
				"start_time": 1,
				"end_time":   1,
				"problems":   1,
				"v_members":  1,
			},
		)
	var contest foundationmodel.ContestViewRank
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest, nil
}

func (d *ContestDao) GetProblems(ctx context.Context, id int) ([]*foundationmodel.ContestProblem, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"problems": 1,
		},
	)
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

func (d *ContestDao) GetProblemIndex(ctx context.Context, id int, problemId string) (int, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"problems": bson.M{
				"$elemMatch": bson.M{
					"problem_id": problemId,
				},
			},
		},
	)
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
		return 0, fmt.Errorf("problem_id %s not found", problemId)
	}
	return result.Problems[0].Index, nil
}

func (d *ContestDao) GetProblemIdByContest(ctx context.Context, id int, problemIndex int) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"problems": bson.M{
				"$elemMatch": bson.M{
					"index": problemIndex,
				},
			},
		},
	)
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

func (d *ContestDao) GetContestOwnerId(ctx context.Context, id int) (int, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":      1,
				"owner_id": 1,
			},
		)
	var contest struct {
		Id      int `bson:"_id"`
		OwnerId int `bson:"owner_id"`
	}
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, metaerror.Wrap(err, "find contest error")
	}
	return contest.OwnerId, nil
}

func (d *ContestDao) GetContestList(
	ctx context.Context,
	page int,
	pageSize int,
) (
	[]*foundationmodel.Contest,
	int,
	error,
) {
	filter := bson.M{}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(
			bson.M{
				"_id":        1,
				"title":      1,
				"start_time": 1,
				"end_time":   1,
				"owner_id":   1,
			},
		).
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

func (d *ContestDao) HasContestSubmitAuth(ctx context.Context, id int, userId int) (bool, error) {
	filter := bson.D{
		{"_id", id},
		{
			"$or", bson.A{
				bson.D{{"owner_id", userId}},
				bson.D{{"private", bson.M{"$exists": false}}},
				bson.D{{"members", bson.M{"$in": []int{userId}}}},
			},
		},
	}
	count, err := d.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *ContestDao) UpdateContest(ctx context.Context, contestId int, contest *foundationmodel.Contest) error {
	filter := bson.D{
		{"_id", contestId},
	}
	setData := metamongo.StructToMapInclude(
		contest,
		"title",
		"description",
		"notification",
		"start_time",
		"end_time",
		"problems",
		"private",
		"members",
		"update_time",
		"lock_rank_duration",
		"always_lock",
	)
	update := bson.M{
		"$set": setData,
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update contest, id: %d", contestId)
	}
	return nil
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
	_, err = sess.WithTransaction(
		ctx, func(sc mongo.SessionContext) (interface{}, error) {
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
		},
	)
	if err != nil {
		return err
	}
	return nil
}
