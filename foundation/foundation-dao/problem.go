package foundationdao

import (
	"context"
	"errors"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
	"regexp"
	"strconv"
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
	_, err := d.collection.Indexes().CreateMany(
		ctx, []mongo.IndexModel{
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
			{
				// 对private字段建立索引，方便查询公开题目
				Keys: bson.D{
					{Key: "private", Value: 1},
				},
				Options: options.Index().SetName("idx_private"),
			},
			{
				// 文本索引，用于全文搜索（title 和 description），但由于中文占比高不太好用
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "description", Value: "text"},
				},
				Options: options.Index().
					SetName("idx_text_search").
					SetWeights(
						bson.D{
							{Key: "title", Value: 10},      // 提高 title 权重
							{Key: "description", Value: 2}, // 降低 description 权重
						},
					),
			},
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to create index for problem collection")
	}
	return nil
}

func (d *ProblemDao) GetProblemEditAuth(ctx context.Context, id string) (*foundationmodel.ProblemViewAuth, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"_id":          1,
			"creator_id":   1,
			"auth_members": 1,
		},
	)
	var problem foundationmodel.ProblemViewAuth
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewAuth(ctx context.Context, id string) (*foundationmodel.ProblemViewAuth, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"_id":          1,
			"creator_id":   1,
			"private":      1,
			"members":      1,
			"auth_members": 1,
		},
	)
	var problem foundationmodel.ProblemViewAuth
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
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

func (d *ProblemDao) HasProblemTitle(ctx context.Context, title string) (bool, error) {
	filter := bson.M{
		"title": title,
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

func (d *ProblemDao) GetProblemView(ctx context.Context, id string, userId int, hasAuth bool) (
	*foundationmodel.Problem,
	error,
) {
	filter := bson.M{
		"_id": id,
	}
	if !hasAuth {
		if userId > 0 {
			filter["$or"] = []bson.M{
				{"private": bson.M{"$exists": false}},
				{"creator_id": userId},
				{"members": userId},
				{"auth_members": userId},
			}
		} else {
			filter["private"] = bson.M{
				"$exists": false,
			}
		}
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"members":      0,
			"auth_members": 0,
		},
	)
	var problem foundationmodel.Problem
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewJudge(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"_id":          1,
			"time_limit":   1,
			"memory_limit": 1,
			"judge_type":   1,
			"judge_md5":    1,
		},
	)
	var problem foundationmodel.Problem
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemTitle(ctx context.Context, id *string) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"_id":   1,
			"title": 1,
		},
	)
	var problem foundationmodel.Problem
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem.Title, nil

}

func (d *ProblemDao) GetProblemTitles(
	ctx context.Context,
	userId int,
	hasAuth bool,
	ids []string,
) ([]*foundationmodel.ProblemViewTitle, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	if !hasAuth {
		if userId > 0 {
			filter["$or"] = []bson.M{
				{"private": bson.M{"$exists": false}},
				{"creator_id": userId},
				{"members": userId},
				{"auth_members": userId},
			}
		} else {
			filter["private"] = bson.M{
				"$exists": false,
			}
		}
	}
	opts := options.Find().SetProjection(
		bson.M{
			"_id":   1,
			"title": 1,
		},
	)
	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find problem titles error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var titles []*foundationmodel.ProblemViewTitle
	for cursor.Next(ctx) {
		var problem foundationmodel.ProblemViewTitle
		if err := cursor.Decode(&problem); err != nil {
			return nil, metaerror.Wrap(err, "decode problem title error")
		}
		titles = append(titles, &problem)
	}
	if err := cursor.Err(); err != nil {
		return nil, metaerror.Wrap(err, "cursor error")
	}
	return titles, nil
}

func (d *ProblemDao) GetProblemJudge(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"_id":              1,
			"title":            1,
			"insert_time":      1,
			"update_time":      1,
			"creator_id":       1,
			"creator_nickname": 1,
			"judge_md5":        1,
			"judge_type":       1,
		},
	)
	var problem foundationmodel.Problem
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemDescription(ctx context.Context, id string) (*string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"description": 1,
		},
	)
	var problem struct {
		Description *string `bson:"description"`
	}
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&problem); err != nil {
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return problem.Description, nil
}

func (d *ProblemDao) GetProblemList(
	ctx context.Context,
	oj string, title string, tags []int, private bool,
	userId int, hasAuth bool,
	page int,
	pageSize int,
) (
	[]*foundationmodel.Problem,
	int,
	error,
) {
	filter := bson.M{}
	if !hasAuth {
		if userId > 0 {
			filter["$or"] = []bson.M{
				{"private": bson.M{"$exists": false}},
				{"creator_id": userId},
				{"members": userId},
				{"auth_members": userId},
			}
		} else {
			filter["private"] = bson.M{
				"$exists": false,
			}
		}
	} else {
		if private {
			filter["private"] = private
		}
	}
	if oj == "didaoj" {
		filter["origin_oj"] = bson.M{
			"$exists": false, // 如果oj为空，则查询没有origin_oj的记录
		}
	} else if oj != "" {
		filter["origin_oj"] = oj
	}
	if title != "" {
		filter["title"] = bson.M{
			"$regex":   regexp.QuoteMeta(title),
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
		SetProjection(
			bson.M{
				"_id":     1,
				"title":   1,
				"tags":    1,
				"accept":  1,
				"attempt": 1,
			},
		).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(
			bson.D{
				{"sort", 1},
				{"_id", 1},
			},
		)
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

func (d *ProblemDao) GetProblemListTitle(ctx context.Context, ids []string) (
	[]*foundationmodel.ProblemViewTitle,
	error,
) {
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

func (d *ProblemDao) GetProblems(ctx context.Context, ids []string) ([]*foundationmodel.Problem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	opts := options.Find().SetProjection(bson.M{"_id": 1, "title": 1, "tags": 1, "accept": 1, "attempt": 1})
	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find problems error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var problems []*foundationmodel.Problem
	for cursor.Next(ctx) {
		var problem foundationmodel.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, metaerror.Wrap(err, "decode problem error")
		}
		problems = append(problems, &problem)
	}
	return problems, nil
}

func (d *ProblemDao) FilterValidProblemIds(ctx *gin.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	opts := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := d.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find problems error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var validIds []string
	for cursor.Next(ctx) {
		var problem foundationmodel.Problem
		if err := cursor.Decode(&problem); err != nil {
			return nil, metaerror.Wrap(err, "decode problem error")
		}
		validIds = append(validIds, problem.Id)
	}
	if err := cursor.Err(); err != nil {
		return nil, metaerror.Wrap(err, "cursor error")
	}
	return validIds, nil
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

func (d *ProblemDao) UpdateProblemsExcludeManualEdit(ctx context.Context, problems []*foundationmodel.Problem) error {
	var models []mongo.WriteModel
	for _, problem := range problems {
		onlyInsertFields := []string{
			"judge_md5",
		}
		setData := metamongo.StructToMapExclude(problem, onlyInsertFields...)
		setInsertData := metamongo.StructToMapInclude(problem, onlyInsertFields...)
		filter := bson.D{
			{"_id", problem.Id},
		}
		update := bson.M{
			"$set":         setData,
			"$setOnInsert": setInsertData,
		}
		updateModel := mongo.NewUpdateManyModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		models = append(models, updateModel)
	}
	bulkOptions := options.BulkWrite().SetOrdered(true) // 设置是否按顺序执行
	_, err := d.collection.BulkWrite(ctx, models, bulkOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to perform bulk update")
	}
	return nil
}

func (d *ProblemDao) InsertProblem(ctx context.Context, problem *foundationmodel.Problem) error {
	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(
		ctx, func(sc mongo.SessionContext) (interface{}, error) {
			seq, err := GetCounterDao().GetNextSequence(sc, "problem_id")
			if err != nil {
				return nil, err
			}
			problem.Id = strconv.Itoa(seq)
			_, err = d.collection.InsertOne(sc, problem)
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

func (d *ProblemDao) PostCreate(ctx context.Context, problem *foundationmodel.Problem, tags []string) (
	*string,
	error,
) {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	var newProblemId *string
	_, err = session.WithTransaction(
		ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			var tagIds []int
			for _, tagName := range tags {
				tag := foundationmodel.NewProblemTagBuilder().Name(tagName).Build()
				err := GetProblemTagDao().InsertTag(sessCtx, tag)
				if err != nil {
					return nil, err
				}
				tagIds = append(tagIds, tag.Id)
			}
			problem.Tags = tagIds
			seq, err := GetCounterDao().GetNextSequence(sessCtx, "problem_id")
			if err != nil {
				return nil, err
			}
			problem.Id = strconv.Itoa(seq)
			problem.Sort = len(problem.Id)
			_, err = d.collection.InsertOne(sessCtx, problem)
			if err != nil {
				return nil, err
			}
			newProblemId = &problem.Id
			return nil, nil
		},
	)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return newProblemId, nil
}

func (d *ProblemDao) UpdateProblem(
	ctx context.Context,
	problemId string,
	problem *foundationmodel.Problem,
	tags []string,
) error {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

			var tagIds []int
			for _, tagName := range tags {
				tag := foundationmodel.NewProblemTagBuilder().Name(tagName).Build()
				err := GetProblemTagDao().InsertTag(sessCtx, tag)
				if err != nil {
					return nil, err
				}
				tagIds = append(tagIds, tag.Id)
			}
			setData := bson.M{
				"title":        problem.Title,
				"description":  problem.Description,
				"tags":         tagIds,
				"update_time":  problem.UpdateTime,
				"time_limit":   problem.TimeLimit,
				"memory_limit": problem.MemoryLimit,
				"source":       problem.Source,
			}
			unsetData := bson.M{}
			if problem.Private {
				setData["private"] = problem.Private
			} else {
				unsetData["private"] = 1
			}
			_, err = d.collection.UpdateOne(
				sessCtx, bson.M{"_id": problemId}, bson.M{
					"$set":   setData,
					"$unset": unsetData,
				},
			)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemDescription(
	ctx context.Context,
	id string,
	description string,
) error {
	nowTime := metatime.GetTimeNow()
	filter := bson.M{
		"_id": id,
	}
	update := bson.M{
		"$set": bson.M{
			"description": description,
			"update_time": nowTime,
		},
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judge md5")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemJudgeInfo(
	ctx context.Context,
	id string,
	judgeType foundationjudge.JudgeType,
	md5 string,
) error {
	nowTime := metatime.GetTimeNow()
	filter := bson.M{
		"_id": id,
	}
	update := bson.M{
		"$set": bson.M{
			"judge_type":  judgeType,
			"judge_md5":   md5,
			"update_time": nowTime,
		},
	}
	_, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judge md5")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemCrawl(ctx context.Context, id string, problem *foundationmodel.Problem) error {
	filter := bson.M{
		"_id": id,
	}
	onlyInsertFields := []string{
		"insert_time",
		"accept",
		"attempt",
	}
	setData := metamongo.StructToMapExclude(problem, onlyInsertFields...)
	setInsertData := metamongo.StructToMapInclude(problem, onlyInsertFields...)
	update := bson.M{
		"$set":         setData,
		"$setOnInsert": setInsertData,
	}
	opts := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return metaerror.Wrap(err, "failed to update problem crawl time")
	}
	return nil
}
