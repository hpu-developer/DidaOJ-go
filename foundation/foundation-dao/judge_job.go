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
	"log/slog"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"time"
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
	collection := d.collection
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "approve_time", Value: 1},
				{Key: "author_id", Value: 1},
				{Key: "problem_id", Value: 1},
			},
			Options: options.Index().SetName("idx_status_author_problem"),
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (d *JudgeJobDao) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
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
			seq, err := GetCounterDao().GetNextSequence(sc, "judge_id")
			if err != nil {
				return nil, err
			}
			// 更新 judgeJob 的 ID
			judgeJob.Id = seq
			// 插入新的 JudgeJob
			_, err = d.collection.InsertOne(sc, judgeJob)
			if err != nil {
				return nil, err
			}
			// 更新Problem表的attempt计数
			problemAttempt := bson.M{
				"attempt": 1,
			}
			if judgeJob.Status == foundationjudge.JudgeStatusAC {
				problemAttempt["accept"] = 1
			}
			_, err = GetProblemDao().collection.UpdateOne(
				sc,
				bson.M{"_id": judgeJob.ProblemId},
				bson.M{"$inc": problemAttempt},
			)
			// 更新User表的attempt计数
			userAttempt := bson.M{
				"attempt": 1,
			}
			if judgeJob.Status == foundationjudge.JudgeStatusAC {
				userAttempt["accept"] = 1
			}
			_, err = GetUserDao().collection.UpdateOne(
				sc,
				bson.M{"_id": judgeJob.AuthorId},
				bson.M{"$inc": userAttempt},
			)

			if err != nil {
				return nil, metaerror.Wrap(
					err,
					"failed to update problem attempt count for problem %s",
					judgeJob.ProblemId,
				)
			}
			return nil, nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *JudgeJobDao) UpdateJudgeJob(ctx context.Context, id int, judgeSource *foundationmodel.JudgeJob) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": judgeSource,
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update job")
	}
	return nil
}

func (d *JudgeJobDao) GetJudgeJob(ctx context.Context, judgeId int) (*foundationmodel.JudgeJob, error) {
	filter := bson.M{
		"_id": judgeId,
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

func (d *JudgeJobDao) GetJudgeJobViewAuth(ctx *gin.Context, id int) (*foundationmodel.JudgeJobViewAuth, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"_id":          1,
				"contest_id":   1,
				"author_id":    1,
				"approve_time": 1,
				"private":      1, // 是否隐藏源码
			},
		)
	var judgeSource foundationmodel.JudgeJobViewAuth
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&judgeSource); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find judgeSource error")
	}
	return &judgeSource, nil
}

func (d *JudgeJobDao) GetJudgeCode(ctx context.Context, id int) (foundationjudge.JudgeLanguage, *string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(
			bson.M{
				"language": 1,
				"code":     1,
			},
		)
	var judgeSource struct {
		Language foundationjudge.JudgeLanguage `bson:"language"`
		Code     *string                       `bson:"code"`
	}
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&judgeSource); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return foundationjudge.JudgeLanguageUnknown, nil, nil
		}
		return foundationjudge.JudgeLanguageUnknown, nil, metaerror.Wrap(err, "find judgeSource error")
	}
	return judgeSource.Language, judgeSource.Code, nil

}

func (d *JudgeJobDao) GetJudgeJobList(
	ctx context.Context,
	contestId int, problemId string,
	searchUserId int, language foundationjudge.JudgeLanguage, status foundationjudge.JudgeStatus,
	page int, pageSize int,
) ([]*foundationmodel.JudgeJob, int, error) {
	filter := bson.M{}
	if contestId > 0 {
		filter["contest_id"] = contestId
	} else {
		filter["contest_id"] = bson.M{"$exists": false}
	}
	if problemId != "" {
		filter["problem_id"] = problemId
	}
	if searchUserId > 0 {
		filter["author_id"] = searchUserId
	}
	if foundationjudge.IsValidJudgeLanguage(int(language)) {
		filter["language"] = language
	}
	if foundationjudge.IsValidJudgeStatus(int(status)) {
		filter["status"] = status
	}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(
			bson.M{
				"_id":          1,
				"approve_time": 1,
				"language":     1,
				"score":        1,
				"status":       1,
				"time":         1,
				"memory":       1,
				"problem_id":   1,
				"author_id":    1,
				"code_length":  1,
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
	var list []*foundationmodel.JudgeJob
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *JudgeJobDao) GetProblemAttemptStatus(
	ctx context.Context, problemIds []string, authorId int, contestId int,
) (map[string]foundationmodel.ProblemAttemptStatus, error) {
	match := bson.M{
		"author_id":  authorId,
		"problem_id": bson.M{"$in": problemIds},
	}
	if contestId > 0 {
		match["contest_id"] = contestId
	}
	pipeline := mongo.Pipeline{
		{
			{
				Key:   "$match",
				Value: match,
			},
		},
		{
			{
				Key: "$group", Value: bson.M{
					"_id": "$problem_id",
					"statusSum": bson.M{
						"$sum": bson.M{
							"$cond": []interface{}{
								bson.M{"$eq": []interface{}{"$status", foundationjudge.JudgeStatusAC}},
								2, // 完成就加2
								1, // 其他状态加1（尝试）
							},
						},
					},
				},
			},
		},
		{
			{
				Key: "$project", Value: bson.M{
					"problem_id": "$_id",
					"finalStatus": bson.M{
						"$cond": bson.A{
							bson.M{"$gte": bson.A{"$statusSum", 2}},
							2, // >=2，有完成记录
							1, // 否则就是尝试过
						},
					},
				},
			},
		},
	}
	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to aggregate judge job")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	type Result struct {
		ProblemId   string                               `bson:"problem_id"`
		FinalStatus foundationmodel.ProblemAttemptStatus `bson:"finalStatus"`
	}
	var results []Result
	if err := cursor.All(ctx, &results); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode aggregation result")
	}
	statusMap := make(map[string]foundationmodel.ProblemAttemptStatus, len(problemIds))
	for _, r := range results {
		statusMap[r.ProblemId] = r.FinalStatus
	}
	return statusMap, nil
}

func (d *JudgeJobDao) GetProblemContestViewAttempt(
	ctx context.Context,
	contestId int,
	problemIds []string,
) ([]*foundationmodel.ProblemViewAttempt, error) {
	pipeline := mongo.Pipeline{
		{
			{
				"$match", bson.D{
					{"contest_id", contestId},
					{"problem_id", bson.D{{"$in", problemIds}}},
				},
			},
		},
		{
			{
				"$group", bson.D{
					{"_id", "$problem_id"}, // 关键修正
					{"attempt", bson.D{{"$sum", 1}}},
					{
						"accept", bson.D{
							{
								"$sum", bson.D{
									{
										"$cond", bson.A{
											bson.D{{"$eq", bson.A{"$status", foundationjudge.JudgeStatusAC}}},
											1,
											0,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	var results []*foundationmodel.ProblemViewAttempt
	for cursor.Next(ctx) {
		var r struct {
			Id      string `bson:"_id"`
			Attempt int    `bson:"attempt"`
			Accept  int    `bson:"accept"`
		}
		if err := cursor.Decode(&r); err != nil {
			return nil, err
		}
		results = append(
			results, &foundationmodel.ProblemViewAttempt{
				Id:      r.Id,
				Attempt: r.Attempt,
				Accept:  r.Accept,
			},
		)
	}
	return results, nil
}

func (d *JudgeJobDao) GetProblemTimeViewAttempt(
	ctx context.Context,
	startTime *time.Time,
	endTime *time.Time,
	problemIds []string,
	members []int,
) ([]*foundationmodel.ProblemViewAttempt, error) {
	match := bson.M{
		"problem_id": bson.M{"$in": problemIds},
		"author_id":  bson.M{"$in": members},
	}

	timeCond := bson.M{}
	if startTime != nil {
		timeCond["$gte"] = startTime
	}
	if endTime != nil {
		timeCond["$lte"] = endTime
	}
	if len(timeCond) > 0 {
		match["approve_time"] = timeCond
	}

	pipeline := mongo.Pipeline{
		{{"$match", match}},
		{
			{
				"$group", bson.M{
					"_id":     "$problem_id",
					"attempt": bson.M{"$sum": 1},
					"accept": bson.M{
						"$sum": bson.M{
							"$cond": bson.A{
								bson.M{"$eq": bson.A{"$status", foundationjudge.JudgeStatusAC}},
								1,
								0,
							},
						},
					},
				},
			},
		},
	}
	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	var results []*foundationmodel.ProblemViewAttempt
	for cursor.Next(ctx) {
		var r struct {
			Id      string `bson:"_id"`
			Attempt int    `bson:"attempt"`
			Accept  int    `bson:"accept"`
		}
		if err := cursor.Decode(&r); err != nil {
			return nil, err
		}
		results = append(
			results, &foundationmodel.ProblemViewAttempt{
				Id:      r.Id,
				Attempt: r.Attempt,
				Accept:  r.Accept,
			},
		)
	}
	return results, nil
}

func (d *JudgeJobDao) GetRankAcProblem(
	ctx *gin.Context,
	approveStartTime *time.Time,
	approveEndTime *time.Time,
	page int,
	pageSize int,
) ([]*foundationmodel.UserRank, int, error) {
	collection := d.collection

	// 生成 $match 条件
	matchCond := bson.M{"status": foundationjudge.JudgeStatusAC}
	if approveStartTime != nil || approveEndTime != nil {
		timeCond := bson.M{}
		if approveStartTime != nil {
			timeCond["$gte"] = approveStartTime
		}
		if approveEndTime != nil {
			timeCond["$lt"] = approveEndTime
		}
		matchCond["approve_time"] = timeCond
	}

	// 主聚合流水线
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchCond}},
		{
			{
				Key: "$group", Value: bson.M{
					"_id": bson.M{
						"author_id":  "$author_id",
						"problem_id": "$problem_id",
					},
				},
			},
		},
		// 再按用户统计通过题数
		{
			{
				Key: "$group", Value: bson.M{
					"_id":   "$_id.author_id",
					"count": bson.M{"$sum": 1},
				},
			},
		},
		// 按通过题数倒序排列
		{{Key: "$sort", Value: bson.M{"count": -1, "_id": 1}}},
		// 分页
		{{Key: "$skip", Value: (page - 1) * pageSize}},
		{{Key: "$limit", Value: pageSize}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)

	// 构建返回值
	var list []*foundationmodel.UserRank
	for cursor.Next(ctx) {
		var result struct {
			ID    int `bson:"_id"`
			Count int `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, 0, err
		}
		list = append(list, foundationmodel.NewUserRankBuilder().Id(result.ID).ProblemCount(result.Count).Build())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}
	totalPipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchCond}},
		{
			{
				Key: "$group", Value: bson.M{
					"_id": bson.M{
						"author_id":  "$author_id",
						"problem_id": "$problem_id",
					},
				},
			},
		},
		{
			{
				Key: "$group", Value: bson.M{
					"_id": "$_id.author_id",
				},
			},
		},
		{{Key: "$count", Value: "total"}},
	}
	totalCursor, err := collection.Aggregate(ctx, totalPipeline)
	if err != nil {
		return list, 0, nil // 返回分页结果，但不报错
	}
	defer func(totalCursor *mongo.Cursor, ctx context.Context) {
		err := totalCursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(totalCursor, ctx)
	var totalResult struct {
		Total int `bson:"total"`
	}
	if totalCursor.Next(ctx) {
		_ = totalCursor.Decode(&totalResult)
	}
	return list, totalResult.Total, nil
}

func (d *JudgeJobDao) GetUserAcProblemIds(ctx context.Context, userId int) ([]string, error) {
	filter := bson.M{
		"author_id": userId,
		"status":    foundationjudge.JudgeStatusAC,
	}
	values, err := d.collection.Distinct(ctx, "problem_id", filter)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to get distinct problem_ids")
	}
	result := make([]string, 0, len(values))
	for _, v := range values {
		if id, ok := v.(string); ok {
			result = append(result, id)
		}
	}
	return result, nil
}

func (d *JudgeJobDao) GetContestRanks(
	ctx context.Context,
	id int,
	startTime time.Time,
	lockTime *time.Time,
	problemMap map[string]int,
) ([]*foundationmodel.ContestRank, error) {

	match := bson.M{
		"contest_id": id,
	}
	if lockTime != nil {
		match["approve_time"] = bson.M{
			"$gte": startTime,
			"$lt":  lockTime,
		}
	} else {
		match["approve_time"] = bson.M{"$gte": startTime}
	}

	pipeline := mongo.Pipeline{
		{
			{
				"$match", match,
			},
		},
		{
			{
				"$group", bson.M{
					"_id": bson.M{
						"author_id":  "$author_id",
						"problem_id": "$problem_id",
					},
					"ac_list": bson.M{
						"$push": bson.M{
							"_id":          "$_id",
							"status":       "$status",
							"approve_time": "$approve_time",
						},
					},
				},
			},
		},
		{
			{
				"$addFields", bson.M{
					"first_ac": bson.M{
						"$first": bson.M{
							"$filter": bson.M{
								"input": "$ac_list",
								"as":    "s",
								"cond":  bson.M{"$eq": bson.A{"$$s.status", foundationjudge.JudgeStatusAC}},
							},
						},
					},
				},
			},
		},
		{
			{
				"$addFields", bson.M{
					"attempt_count": bson.M{
						"$cond": bson.A{
							bson.M{"$ifNull": bson.A{"$first_ac._id", false}},
							bson.M{
								"$size": bson.M{
									"$filter": bson.M{
										"input": "$ac_list",
										"as":    "s",
										"cond":  bson.M{"$lt": bson.A{"$$s._id", "$first_ac._id"}},
									},
								},
							},
							bson.M{"$size": "$ac_list"},
						},
					},
				},
			},
		},
		{
			{
				"$project", bson.M{
					"author_id":     "$_id.author_id",
					"problem_id":    "$_id.problem_id",
					"first_ac_time": "$first_ac.approve_time",
					"attempt_count": 1,
				},
			},
		},
	}

	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}()

	type entry struct {
		AuthorId     int        `bson:"author_id"`
		ProblemId    string     `bson:"problem_id"`
		FirstAcTime  *time.Time `bson:"first_ac_time"`
		AttemptCount int        `bson:"attempt_count"`
	}

	getProblemIndex := func(problemId string) int {
		if index, ok := problemMap[problemId]; ok {
			return index
		}
		return -1
	}

	rankMap := make(map[int]*foundationmodel.ContestRank)

	for cursor.Next(ctx) {
		var e entry
		if err := cursor.Decode(&e); err != nil {
			return nil, err
		}

		if _, ok := rankMap[e.AuthorId]; !ok {
			rankMap[e.AuthorId] = &foundationmodel.ContestRank{
				AuthorId: e.AuthorId,
			}
		}

		rankMap[e.AuthorId].Problems = append(
			rankMap[e.AuthorId].Problems, &foundationmodel.ContestRankProblem{
				Index:   getProblemIndex(e.ProblemId),
				Ac:      e.FirstAcTime,
				Attempt: e.AttemptCount,
			},
		)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	result := make([]*foundationmodel.ContestRank, 0, len(rankMap))
	for _, v := range rankMap {
		result = append(result, v)
	}

	return result, nil
}

func (d *JudgeJobDao) GetAcceptedProblemCount(
	ctx context.Context,
	startTime *time.Time,
	endTime *time.Time,
	problemIds []string,
	userIds []int,
) (map[int]int, error) {
	match := bson.M{
		"author_id":  bson.M{"$in": userIds},
		"problem_id": bson.M{"$in": problemIds},
		"status":     foundationjudge.JudgeStatusAC, // 只要AC记录
	}
	if startTime != nil || endTime != nil {
		timeCond := bson.M{}
		if startTime != nil {
			timeCond["$gte"] = startTime
		}
		if endTime != nil {
			timeCond["$lte"] = endTime
		}
		match["approve_time"] = timeCond
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			// 每个用户对每道AC题保留一条
			{
				"$group", bson.M{
					"_id": bson.M{
						"author_id":  "$author_id",
						"problem_id": "$problem_id",
					},
				},
			},
		},
		{
			// 再按用户聚合计数
			{
				"$group", bson.M{
					"_id":    "$_id.author_id",
					"accept": bson.M{"$sum": 1},
				},
			},
		},
	}

	cursor, err := d.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)

	type result struct {
		AuthorId int `bson:"_id"`
		Accept   int `bson:"accept"`
	}
	resMap := make(map[int]int)
	for cursor.Next(ctx) {
		var r result
		if err := cursor.Decode(&r); err != nil {
			return nil, err
		}
		resMap[r.AuthorId] = r.Accept
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return resMap, nil
}

func (d *JudgeJobDao) GetProblemRecommendByUser(ctx context.Context, userId int, hasAuth bool) ([]string, error) {
	return d.GetProblemRecommendByProblem(ctx, userId, hasAuth, "")
}

func (d *JudgeJobDao) GetProblemRecommendByProblem(
	ctx context.Context,
	userId int,
	hasAuth bool,
	problemId string,
) ([]string, error) {
	collection := d.collection

	userAcProblems, err := d.GetUserAcProblemIds(ctx, userId)
	if err != nil {
		return nil, err
	}

	match := bson.M{
		"status": foundationjudge.JudgeStatusAC,
	}
	if problemId != "" {
		match["problem_id"] = problemId // 只看当前题
	}

	// Step 1: 找出做过该题的用户
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{
			{
				Key: "$group", Value: bson.M{
					"_id": "$author_id",
				},
			},
		},
		{{Key: "$limit", Value: 1000}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	var acUserIDs []int
	for cursor.Next(ctx) {
		var result struct {
			ID int `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		acUserIDs = append(acUserIDs, result.ID)
	}

	// Step 2: 找出这些用户做过的其他题（排除当前题）
	pipeline = mongo.Pipeline{
		// 第一步：筛选评测记录
		{
			{
				Key: "$match", Value: bson.M{
					"status":       foundationjudge.JudgeStatusAC,
					"approve_time": bson.M{"$exists": true},
					"author_id":    bson.M{"$in": acUserIDs},
					"problem_id":   bson.M{"$nin": userAcProblems},
				},
			},
		},
		// 第二步：分组
		{
			{
				Key: "$group", Value: bson.M{
					"_id":   "$problem_id",
					"count": bson.M{"$sum": 1},
				},
			},
		},
		// 第三步：关联 problem 信息
		{
			{
				Key: "$lookup", Value: bson.M{
					"from":         "problem",
					"localField":   "_id",
					"foreignField": "_id",
					"as":           "look_problem",
				},
			},
		},
		// 第四步：展开关联的 problem（通常每个 problem_id 只对应一个问题）
		{{Key: "$unwind", Value: "$look_problem"}},
	}
	// 过滤掉没有权限的题目
	if !hasAuth {
		var filter bson.M
		if userId > 0 {
			filter = bson.M{
				"$or": []bson.M{
					{"look_problem.private": bson.M{"$exists": false}},
					{"look_problem.auth_users": userId},
				},
			}
		} else {
			filter = bson.M{
				"look_problem.private": bson.M{"$exists": false},
			}
		}
		pipeline = append(
			pipeline, bson.D{
				{Key: "$match", Value: filter},
			},
		)
	}
	pipeline = append(
		pipeline,
		bson.D{{Key: "$sort", Value: bson.M{"count": -1}}},
		bson.D{{Key: "$limit", Value: 20}},
	)

	cursor, err = collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)

	var results []struct {
		ProblemID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	problemIDs := make([]string, 0, len(results))
	for _, r := range results {
		problemIDs = append(problemIDs, r.ProblemID)
	}
	return problemIDs, nil
}

// RequestJudgeJobListPendingJudge 获取待评测的 JudgeJob 列表，优先取最小的
func (d *JudgeJobDao) RequestJudgeJobListPendingJudge(
	ctx context.Context,
	maxCount int,
	judger string,
) ([]*foundationmodel.JudgeJob, error) {
	var result []*foundationmodel.JudgeJob

	for i := 0; i < maxCount; i++ {
		filter := bson.M{
			"status": bson.M{
				"$in": []foundationjudge.JudgeStatus{
					foundationjudge.JudgeStatusInit,
					foundationjudge.JudgeStatusRejudge,
				},
			},
		}

		update := bson.M{
			"$set": bson.M{
				"status":     foundationjudge.JudgeStatusQueuing, // 标记为处理中
				"judger":     judger,
				"judge_time": time.Now(), // 标记开始处理的时间，可以据此判断重试
			},
		}

		findOpts := options.FindOneAndUpdate().
			SetSort(bson.M{"_id": 1}).
			SetReturnDocument(options.After) // 返回更新后的文档

		var job foundationmodel.JudgeJob
		err := d.collection.FindOneAndUpdate(ctx, filter, update, findOpts).Decode(&job)
		if errors.Is(err, mongo.ErrNoDocuments) {
			break // 没有更多了
		}
		if err != nil {
			return nil, metaerror.Wrap(err, "findOneAndUpdate error")
		}
		result = append(result, &job)
	}

	return result, nil
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

func (d *JudgeJobDao) StartProcessJudgeJob(ctx context.Context, id int, judger string) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": bson.M{
			"status": foundationjudge.JudgeStatusCompiling,
		},
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update job")
	}
	return nil
}

func (d *JudgeJobDao) MarkJudgeJobJudgeStatus(ctx context.Context, id int, status foundationjudge.JudgeStatus) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to mark judge job status")
	}
	return nil
}

func (d *JudgeJobDao) MarkJudgeJobJudgeFinalStatus(
	ctx context.Context, id int,
	status foundationjudge.JudgeStatus,
	problemId string,
	userId int,
	score int,
	time int,
	memory int,
) error {

	markStatusFunc := func(ctx context.Context, id int, status foundationjudge.JudgeStatus) error {
		filter := bson.D{
			{"_id", id},
		}
		update := bson.M{
			"$set": bson.M{
				"status": status,
				"score":  score,
				"time":   time,
				"memory": memory,
			},
		}
		updateOptions := options.Update()
		_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
		if err != nil {
			return metaerror.Wrap(err, "failed to mark judge job status")
		}
		return nil
	}

	if status == foundationjudge.JudgeStatusAC {
		session, err := d.collection.Database().Client().StartSession()
		if err != nil {
			return metaerror.Wrap(err, "failed to start mongo session")
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(
			ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
				err := markStatusFunc(sessCtx, id, status)
				if err != nil {
					return nil, err
				}
				// 更新 Problem 表的 accept 计数
				_, err = GetProblemDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": problemId},
					bson.M{
						"$inc": bson.M{
							"accept": 1,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", id)
				}
				// 更新User表的 accept 计数
				_, err = GetUserDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": userId},
					bson.M{
						"$inc": bson.M{
							"accept": 1,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", id)
				}

				return nil, nil
			},
		)
	} else {
		err := markStatusFunc(ctx, id, status)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *JudgeJobDao) MarkJudgeJobCompileMessage(ctx context.Context, id int, message string) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": bson.M{
			"compile_message": message,
		},
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update job")
	}
	return nil
}

func (d *JudgeJobDao) MarkJudgeJobTaskTotal(ctx context.Context, id int, taskTotalCount int) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$set": bson.M{
			"task_current": 0,
			"task_total":   taskTotalCount,
		},
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update job")
	}
	return nil
}

func (d *JudgeJobDao) AddJudgeJobTaskCurrent(ctx context.Context, id int, task *foundationmodel.JudgeTask) error {
	filter := bson.D{
		{"_id", id},
	}
	update := bson.M{
		"$inc": bson.M{
			"task_current": 1,
		},
		"$push": bson.M{
			"task": task,
		},
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update job")
	}
	return nil
}

func (d *JudgeJobDao) RejudgeJob(ctx context.Context, id int) error {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// 1. 查找最近提交
			findFilter := bson.D{{"_id", id}}
			findOpts := options.FindOne().
				SetProjection(
					bson.M{
						"_id":        1,
						"problem_id": 1,
						"author_id":  1,
						"status":     1,
					},
				)
			var doc struct {
				ID        int                         `bson:"_id"`
				ProblemID string                      `bson:"problem_id"`
				AuthorId  int                         `bson:"author_id"`
				Status    foundationjudge.JudgeStatus `bson:"status"`
			}
			if err := d.collection.FindOne(ctx, findFilter, findOpts).Decode(&doc); err != nil {
				return nil, metaerror.Wrap(err, "find judgeSource error")
			}
			problemAcceptDelta := 0
			userAcceptDelta := 0
			if doc.Status == foundationjudge.JudgeStatusAC {
				problemAcceptDelta--
				userAcceptDelta--
			}

			update := bson.M{
				"$set": bson.M{
					"status": foundationjudge.JudgeStatusRejudge,
				},
				"$unset": bson.M{
					"score":           "",
					"time":            "",
					"memory":          "",
					"compile_message": "",
					"task":            "",
					"task_current":    "",
					"task_total":      "",
					"judger":          "",
					"judge_time":      "",
				},
			}
			_, err = d.collection.UpdateOne(sessCtx, findFilter, update)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update submissions")
			}
			// 3. 批量更新 Problem 表的 accept 计数
			if problemAcceptDelta != 0 {
				pid := doc.ProblemID
				_, err := GetProblemDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": pid},
					bson.M{
						"$inc": bson.M{
							"accept": problemAcceptDelta,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", pid)
				}
			}

			// 4. 批量更新 UserId 表的 accept 计数
			if userAcceptDelta != 0 {
				uid := doc.AuthorId
				_, err := GetUserDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": uid},
					bson.M{
						"$inc": bson.M{
							"accept": userAcceptDelta,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", uid)
				}
			}

			return nil, nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return nil
}

func (d *JudgeJobDao) RejudgeSearch(
	ctx context.Context,
	problemId string,
	language foundationjudge.JudgeLanguage,
	status foundationjudge.JudgeStatus,
) error {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// 1. 查找提交
			findFilter := bson.M{}
			if problemId != "" {
				findFilter["problem_id"] = problemId
			}
			if foundationjudge.IsValidJudgeLanguage(int(language)) {
				findFilter["language"] = language
			}
			if foundationjudge.IsValidJudgeStatus(int(status)) {
				findFilter["status"] = status
			}
			findOpts := options.Find().
				SetSort(bson.D{{"_id", -1}}).
				SetProjection(
					bson.M{
						"_id":        1,
						"problem_id": 1,
						"author_id":  1,
						"status":     1,
					},
				)
			cursor, err := d.collection.Find(sessCtx, findFilter, findOpts)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to find recent submissions")
			}
			defer func(cursor *mongo.Cursor, ctx context.Context) {
				err := cursor.Close(ctx)
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
				}
			}(cursor, sessCtx)

			var ids []int
			problemAcceptDelta := map[string]int{} // problem_id => acceptDelta
			userAcceptDelta := map[int]int{}       // user_id => acceptDelta

			for cursor.Next(sessCtx) {
				var doc struct {
					ID        int                         `bson:"_id"`
					ProblemID string                      `bson:"problem_id"`
					AuthorId  int                         `bson:"author_id"`
					Status    foundationjudge.JudgeStatus `bson:"status"`
				}
				if err := cursor.Decode(&doc); err != nil {
					return nil, metaerror.Wrap(err, "failed to decode document")
				}
				ids = append(ids, doc.ID)

				if doc.Status == foundationjudge.JudgeStatusAC {
					problemAcceptDelta[doc.ProblemID]--
					userAcceptDelta[doc.AuthorId]--
				}
			}
			if err := cursor.Err(); err != nil {
				return nil, metaerror.Wrap(err, "cursor error")
			}
			if len(ids) == 0 {
				return nil, nil
			}

			// 2. 批量更新提交状态
			filter := bson.M{"_id": bson.M{"$in": ids}}
			update := bson.M{
				"$set": bson.M{
					"status": foundationjudge.JudgeStatusRejudge,
				},
				"$unset": bson.M{
					"score":           "",
					"time":            "",
					"memory":          "",
					"compile_message": "",
					"task":            "",
					"task_current":    "",
					"task_total":      "",
					"judger":          "",
					"judge_time":      "",
				},
			}
			_, err = d.collection.UpdateMany(sessCtx, filter, update)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update submissions")
			}

			// 3. 批量更新 Problem 表的 accept 计数
			for pid, acceptDelta := range problemAcceptDelta {
				_, err := GetProblemDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": pid},
					bson.M{
						"$inc": bson.M{
							"accept": acceptDelta,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", pid)
				}
			}

			// 4. 批量更新 UserId 表的 accept 计数
			for uid, acceptDelta := range userAcceptDelta {
				_, err := GetUserDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": uid},
					bson.M{
						"$inc": bson.M{
							"accept": acceptDelta,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", uid)
				}
			}

			return nil, nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return nil
}

func (d *JudgeJobDao) RejudgeRecently(ctx context.Context) error {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// 1. 查找最近提交
			findFilter := bson.D{
				{"origin_oj", bson.M{"$exists": false}},
				{
					"$or", bson.A{
						bson.M{"origin_oj": ""},
						bson.M{"origin_oj": nil},
					},
				},
			}
			findOpts := options.Find().
				SetSort(bson.D{{"_id", -1}}).
				SetProjection(
					bson.M{
						"_id":        1,
						"problem_id": 1,
						"author_id":  1,
						"status":     1,
					},
				).
				SetLimit(100)
			cursor, err := d.collection.Find(sessCtx, findFilter, findOpts)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to find recent submissions")
			}
			defer func(cursor *mongo.Cursor, ctx context.Context) {
				err := cursor.Close(ctx)
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
				}
			}(cursor, sessCtx)

			var ids []int
			problemAcceptDelta := map[string]int{} // problem_id => acceptDelta
			userAcceptDelta := map[int]int{}       // user_id => acceptDelta

			for cursor.Next(sessCtx) {
				var doc struct {
					ID        int                         `bson:"_id"`
					ProblemID string                      `bson:"problem_id"`
					AuthorId  int                         `bson:"author_id"`
					Status    foundationjudge.JudgeStatus `bson:"status"`
				}
				if err := cursor.Decode(&doc); err != nil {
					return nil, metaerror.Wrap(err, "failed to decode document")
				}
				ids = append(ids, doc.ID)

				if doc.Status == foundationjudge.JudgeStatusAC {
					problemAcceptDelta[doc.ProblemID]--
					userAcceptDelta[doc.AuthorId]--
				}
			}
			if err := cursor.Err(); err != nil {
				return nil, metaerror.Wrap(err, "cursor error")
			}
			if len(ids) == 0 {
				return nil, nil
			}

			// 2. 批量更新提交状态
			filter := bson.M{"_id": bson.M{"$in": ids}}
			update := bson.M{
				"$set": bson.M{
					"status": foundationjudge.JudgeStatusRejudge,
				},
				"$unset": bson.M{
					"score":           "",
					"time":            "",
					"memory":          "",
					"compile_message": "",
					"task":            "",
					"task_current":    "",
					"task_total":      "",
					"judger":          "",
					"judge_time":      "",
				},
			}
			_, err = d.collection.UpdateMany(sessCtx, filter, update)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update submissions")
			}

			// 3. 批量更新 Problem 表的 accept 计数
			for pid, acceptDelta := range problemAcceptDelta {
				_, err := GetProblemDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": pid},
					bson.M{
						"$inc": bson.M{
							"accept": acceptDelta,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", pid)
				}
			}

			// 4. 批量更新 UserId 表的 accept 计数
			for uid, acceptDelta := range userAcceptDelta {
				_, err := GetUserDao().collection.UpdateOne(
					sessCtx,
					bson.M{"_id": uid},
					bson.M{
						"$inc": bson.M{
							"accept": acceptDelta,
						},
					},
				)
				if err != nil {
					return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", uid)
				}
			}

			return nil, nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return nil
}

func (d *JudgeJobDao) RejudgeAll(ctx context.Context) error {
	const pageSize = 1000
	var err error
	var lastID int
	for {
		lastID, err = d.rejudgeAllChunk(ctx, lastID, pageSize)
		if err != nil {
			return metaerror.Wrap(err, "failed to rejudge all submissions")
		}
		if lastID < 0 {
			break
		}
	}
	return nil
}

func (d *JudgeJobDao) rejudgeAllChunk(ctx context.Context, lastID int, pageSize int) (int, error) {

	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return -1, metaerror.Wrap(err, "failed to start mongo session")
	}

	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// 构造分页 filter（根据 lastID）

			// 如果存在远程OJ，则这里不进行重判，避免对远端OJ造成干扰
			findFilter := bson.M{
				"$or": bson.A{
					bson.M{"origin_oj": bson.M{"$exists": false}},
					bson.M{"origin_oj": nil},
					bson.M{"origin_oj": ""},
				},
			}
			if lastID > 0 {
				findFilter["_id"] = bson.M{"$gt": lastID}
			}

			findOpts := options.Find().
				SetSort(bson.D{{"_id", 1}}).
				SetLimit(int64(pageSize)).
				SetProjection(
					bson.M{
						"_id":        1,
						"problem_id": 1,
						"author_id":  1,
						"status":     1,
					},
				)

			cursor, err := d.collection.Find(sessCtx, findFilter, findOpts)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to find submissions")
			}
			defer func(cursor *mongo.Cursor, ctx context.Context) {
				err := cursor.Close(ctx)
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
				}
			}(cursor, sessCtx)

			var (
				ids                []int
				problemAcceptDelta = map[string]int{}
				userAcceptDelta    = map[int]int{}
				lastDocID          int
			)

			for cursor.Next(sessCtx) {
				var doc struct {
					ID        int                         `bson:"_id"`
					ProblemID string                      `bson:"problem_id"`
					AuthorId  int                         `bson:"author_id"`
					Status    foundationjudge.JudgeStatus `bson:"status"`
				}
				if err := cursor.Decode(&doc); err != nil {
					return nil, metaerror.Wrap(err, "decode error")
				}
				ids = append(ids, doc.ID)
				lastDocID = doc.ID // 保留最后一项的 ID

				if doc.Status == foundationjudge.JudgeStatusAC {
					problemAcceptDelta[doc.ProblemID]--
					userAcceptDelta[doc.AuthorId]--
				}
			}

			if err := cursor.Err(); err != nil {
				return nil, metaerror.Wrap(err, "cursor error")
			}
			if len(ids) == 0 {
				return nil, mongo.ErrNoDocuments
			}

			// 批量更新 JudgeJob
			filter := bson.M{"_id": bson.M{"$in": ids}}
			update := bson.M{
				"$set": bson.M{
					"status": foundationjudge.JudgeStatusRejudge,
				},
				"$unset": bson.M{
					"score":           "",
					"time":            "",
					"memory":          "",
					"compile_message": "",
					"task":            "",
					"task_current":    "",
					"task_total":      "",
					"judger":          "",
					"judge_time":      "",
				},
			}
			if _, err := d.collection.UpdateMany(sessCtx, filter, update); err != nil {
				return nil, metaerror.Wrap(err, "failed to update submissions")
			}

			// 更新 Problem 表
			if len(problemAcceptDelta) > 0 {
				var models []mongo.WriteModel
				for pid, delta := range problemAcceptDelta {
					models = append(
						models, mongo.NewUpdateOneModel().
							SetFilter(bson.M{"_id": pid}).
							SetUpdate(bson.M{"$inc": bson.M{"accept": delta}}),
					)
				}
				if _, err := GetProblemDao().collection.BulkWrite(sessCtx, models); err != nil {
					return nil, metaerror.Wrap(err, "update problem fail")
				}
			}

			// 更新 User 表
			if len(userAcceptDelta) > 0 {
				var models []mongo.WriteModel
				for uid, delta := range userAcceptDelta {
					models = append(
						models, mongo.NewUpdateOneModel().
							SetFilter(bson.M{"_id": uid}).
							SetUpdate(bson.M{"$inc": bson.M{"accept": delta}}),
					)
				}
				if _, err := GetUserDao().collection.BulkWrite(sessCtx, models); err != nil {
					return nil, metaerror.Wrap(err, "update user fail")
				}
			}

			// 成功处理一页
			lastID = lastDocID
			slog.Info("RejudgePage success", "count", len(ids), "lastID", lastID)
			return nil, nil
		},
	)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return -1, nil
	}
	if err != nil {
		return lastID, metaerror.Wrap(err, "transaction failed")
	}

	return lastID, nil
}
