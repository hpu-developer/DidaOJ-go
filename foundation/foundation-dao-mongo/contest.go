package foundationdaomongo

import (
	"context"
	"errors"
	"fmt"
	foundationmodel "foundation/foundation-model-mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
	"regexp"
	"time"
)

type ContestDao struct {
	contest *mongo.Collection
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
			ContestDao.contest = client.
				Database("didaoj").
				Collection("contest")
			return &ContestDao
		},
	)
}

func (d *ContestDao) InitDao(ctx context.Context) error {

	_, err := d.contest.Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "title", Value: 1},
				},
				Options: options.Index().SetName("title"),
			},
			{
				Keys: bson.D{
					{Key: "owner_id", Value: 1},
				},
				Options: options.Index().SetName("owner_id"),
			},
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to create indexes for contest contest")
	}
	return nil
}

func (d *ContestDao) HasContestTitle(ctx context.Context, ownerId int, title string) (bool, error) {
	filter := bson.M{
		"title":    title,
		"owner_id": bson.M{"$ne": ownerId},
	}
	count, err := d.contest.CountDocuments(ctx, filter)
	if err != nil {
		return false, metaerror.Wrap(err, "failed to count documents")
	}
	return count > 0, nil
}

func (d *ContestDao) GetContestDescription(ctx context.Context, id int) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":         1,
				"description": 1,
			},
		)
	var contest struct {
		Description string `bson:"description"`
	}
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest description error")
	}
	return &contest.Description, nil
}

func (d *ContestDao) GetListAll(ctx context.Context) ([]*foundationmodel.Contest, error) {
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := d.contest.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find all contests error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var contests []*foundationmodel.Contest
	for cursor.Next(ctx) {
		var contest foundationmodel.Contest
		if err := cursor.Decode(&contest); err != nil {
			return nil, metaerror.Wrap(err, "decode contest error")
		}
		contests = append(contests, &contest)
	}
	return contests, nil
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
				"submit_anytime":     1,
				"password":           1,
			},
		)
	var contest foundationmodel.Contest
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
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
				"password":           1,
				"type":               1,
				"score_type":         1,
				"descriptions":       1,
				"notification":       1,
				"members":            1,
				"lock_rank_duration": 1,
				"always_lock":        1,
				"submit_anytime":     1,
			},
		)
	var contest foundationmodel.Contest
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
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
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
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
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
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
				"_id":                1,
				"start_time":         1,
				"end_time":           1,
				"problems":           1,
				"v_members":          1,
				"lock_rank_duration": 1,
				"always_lock":        1,
			},
		)
	var contest foundationmodel.ContestViewRank
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
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
	err := d.contest.FindOne(ctx, filter, opts).Decode(&result)
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
	err := d.contest.FindOne(ctx, filter, opts).Decode(&result)
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
	err := d.contest.FindOne(ctx, filter, opts).Decode(&result)
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
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, metaerror.Wrap(err, "find contest error")
	}
	return contest.OwnerId, nil
}

func (d *ContestDao) GetContestStartTime(ctx context.Context, id int) (*time.Time, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":        1,
				"start_time": 1,
			},
		)
	var contest struct {
		StartTime time.Time `bson:"start_time"`
	}
	if err := d.contest.FindOne(ctx, filter, opts).Decode(&contest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest.StartTime, nil
}

func (d *ContestDao) GetContestList(
	ctx context.Context,
	title string,
	userId int,
	page int,
	pageSize int,
) ([]*foundationmodel.Contest, int, error) {
	filter := bson.M{}
	if title != "" {
		filter["title"] = bson.M{
			"$regex":   regexp.QuoteMeta(title),
			"$options": "i", // 不区分大小写
		}
	}
	if userId > 0 {
		filter["owner_id"] = userId
	}
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
				"private":    1,
			},
		).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(
			bson.D{
				{"start_time", -1},
				{"_id", -1},
			},
		)
	// 查询总记录数
	totalCount, err := d.contest.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count documents, page: %d", page)
	}
	// 查询当前页的数据
	cursor, err := d.contest.Find(ctx, filter, opts)
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

func (d *ContestDao) HasContestViewAuth(ctx context.Context, id int, userId int) (bool, error) {
	filter := bson.D{
		{"_id", id},
		{
			"$or", bson.A{
				bson.D{{"owner_id", userId}},
				bson.D{{"private", bson.M{"$exists": false}}},
				bson.D{{"members", bson.M{"$in": []int{userId}}}},
			},
		},
		{"start_time", bson.M{"$lte": metatime.GetTimeNow()}},
	}
	count, err := d.contest.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *ContestDao) HasContestSubmitAuth(ctx context.Context, id int, userId int) (bool, error) {
	nowTime := metatime.GetTimeNow()
	filter := bson.D{
		{"_id", id},
		{
			"$or", bson.A{
				bson.D{{"owner_id", userId}},
				bson.D{{"private", bson.M{"$exists": false}}},
				bson.D{{"members", bson.M{"$in": []int{userId}}}},
			},
		},
		{
			"$or", bson.A{
				bson.D{{"submit_anytime", true}},
				bson.D{{"end_time", bson.M{"$gte": nowTime}}},
			},
		},
		{
			"start_time", bson.M{"$lte": nowTime},
		},
	}
	count, err := d.contest.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *ContestDao) UpdateDescription(ctx context.Context, id int, description string) error {
	filter := bson.M{
		"_id": id,
	}
	update := bson.M{
		"$set": bson.M{
			"description": description,
		},
	}
	_, err := d.contest.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update contest description, id: %d", id)
	}
	return nil
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
		"members",
		"update_time",
		"lock_rank_duration",
		"always_lock",
	)
	unsetData := bson.M{}

	if contest.Private {
		setData["private"] = true
	} else {
		unsetData["private"] = 1
	}
	if contest.SubmitAnytime {
		setData["submit_anytime"] = true
	} else {
		unsetData["submit_anytime"] = 1
	}
	if contest.Password != nil {
		setData["password"] = contest.Password
	} else {
		unsetData["password"] = 1
	}
	update := bson.M{
		"$set":   setData,
		"$unset": unsetData,
	}
	_, err := d.contest.UpdateOne(ctx, filter, update)
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
	_, err := d.contest.BulkWrite(ctx, models, bulkOptions)
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
			_, err = d.contest.InsertOne(sc, contest)
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

func (d *ContestDao) PostPassword(ctx context.Context, userId int, contestId int, password string) (bool, error) {
	filter := bson.M{
		"_id":      contestId,
		"password": password,
	}
	update := bson.M{
		"$addToSet": bson.M{
			"members": userId,
		},
	}
	res, err := d.contest.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, metaerror.Wrap(err, "failed to post contest password, contestId: %d", contestId)
	}
	if res.MatchedCount == 0 {
		return false, nil
	}
	return true, nil
}
