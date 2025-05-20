package foundationdao

import (
	"context"
	"errors"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
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
	return nil
}

func (d *JudgeJobDao) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
	mongoSubsystem := metamongo.GetSubsystem()
	client := mongoSubsystem.GetClient()
	sess, err := client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
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
		_, err = GetProblemDao().collection.UpdateOne(sc,
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
		_, err = GetUserDao().collection.UpdateOne(sc,
			bson.M{"_id": judgeJob.AuthorId},
			bson.M{"$inc": userAttempt},
		)

		if err != nil {
			return nil, metaerror.Wrap(err, "failed to update problem attempt count for problem %s", judgeJob.ProblemId)
		}
		return nil, nil
	})
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

func (d *JudgeJobDao) GetJudgeCode(ctx context.Context, id int) (foundationjudge.JudgeLanguage, *string, error) {
	filter := bson.M{
		"_id": id,
	}
	opts := options.FindOne().
		SetProjection(bson.M{
			"language": 1,
			"code":     1,
		})
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

func (d *JudgeJobDao) GetJudgeJobList(ctx context.Context,
	problemId string, userId int, language foundationjudge.JudgeLanguage, status foundationjudge.JudgeStatus,
	page int, pageSize int,
) ([]*foundationmodel.JudgeJob, int, error) {
	filter := bson.M{}
	if problemId != "" {
		filter["problem_id"] = problemId
	}
	if userId > 0 {
		filter["author_id"] = userId
	}
	if language != foundationjudge.JudgeLanguageUnknown {
		filter["language"] = language
	}
	if status != foundationjudge.JudgeStatusUnknown {
		filter["status"] = status
	}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	// 只获取id、title、tags、accept
	opts := options.Find().
		SetProjection(bson.M{
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
	var list []*foundationmodel.JudgeJob
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *JudgeJobDao) GetProblemAttemptStatus(ctx context.Context, problemIds []string, authorId int,
) (map[string]foundationmodel.ProblemAttemptStatus, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"author_id":  authorId,
			"problem_id": bson.M{"$in": problemIds},
		}}},
		{{Key: "$group", Value: bson.M{
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
		}}},
		{{Key: "$project", Value: bson.M{
			"problem_id": "$_id",
			"finalStatus": bson.M{
				"$cond": bson.A{
					bson.M{"$gte": bson.A{"$statusSum", 2}},
					2, // >=2，有完成记录
					1, // 否则就是尝试过
				},
			},
		}}},
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

func (d *JudgeJobDao) GetProblemViewAttempt(
	ctx context.Context,
	contestId int,
	problemIds []string,
) ([]*foundationmodel.ProblemViewAttempt, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"contest_id", contestId},
			{"problem_id", bson.D{{"$in", problemIds}}},
		}}},
		{{"$group", bson.D{
			{"_id", "$problem_id"}, // 关键修正
			{"attempt", bson.D{{"$sum", 1}}},
			{"accept", bson.D{{"$sum", bson.D{
				{"$cond", bson.A{
					bson.D{{"$eq", bson.A{"$status", foundationjudge.JudgeStatusAC}}},
					1,
					0,
				}},
			}}}},
		}}},
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
		results = append(results, &foundationmodel.ProblemViewAttempt{
			Id:      r.Id,
			Attempt: r.Attempt,
			Accept:  r.Accept,
		})
	}
	return results, nil
}

// RequestJudgeJobListPendingJudge 获取待评测的 JudgeJob 列表，优先取最小的
func (d *JudgeJobDao) RequestJudgeJobListPendingJudge(ctx context.Context, maxCount int, judger string) ([]*foundationmodel.JudgeJob, error) {
	var result []*foundationmodel.JudgeJob

	for i := 0; i < maxCount; i++ {
		filter := bson.M{
			"status": bson.M{"$in": []foundationjudge.JudgeStatus{
				foundationjudge.JudgeStatusInit,
				foundationjudge.JudgeStatusRejudge,
			}},
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

func (d *JudgeJobDao) MarkJudgeJobJudgeFinalStatus(ctx context.Context, id int,
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

		_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			err := markStatusFunc(sessCtx, id, status)
			if err != nil {
				return nil, err
			}
			// 更新 Problem 表的 accept 计数
			_, err = GetProblemDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": problemId},
				bson.M{"$inc": bson.M{
					"accept": 1,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", id)
			}
			// 更新User表的 accept 计数
			_, err = GetUserDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": userId},
				bson.M{"$inc": bson.M{
					"accept": 1,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", id)
			}

			return nil, nil
		})
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

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. 查找最近提交
		findFilter := bson.D{{"_id", id}}
		findOpts := options.FindOne().
			SetProjection(bson.M{
				"_id":        1,
				"problem_id": 1,
				"author_id":  1,
				"status":     1,
			})
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
			_, err := GetProblemDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": pid},
				bson.M{"$inc": bson.M{
					"accept": problemAcceptDelta,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", pid)
			}
		}

		// 4. 批量更新 UserId 表的 accept 计数
		if userAcceptDelta != 0 {
			uid := doc.AuthorId
			_, err := GetUserDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": uid},
				bson.M{"$inc": bson.M{
					"accept": userAcceptDelta,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", uid)
			}
		}

		return nil, nil
	})
	if err != nil {
		return metaerror.Wrap(err, "failed to rejudge submissions in transaction")
	}
	return nil
}

func (d *JudgeJobDao) RejudgeProblem(ctx context.Context, id string) error {
	session, err := d.collection.Database().Client().StartSession()
	if err != nil {
		return metaerror.Wrap(err, "failed to start mongo session")
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. 查找最近提交
		findFilter := bson.D{{"problem_id", id}}
		findOpts := options.Find().
			SetSort(bson.D{{"_id", -1}}).
			SetProjection(bson.M{
				"_id":        1,
				"problem_id": 1,
				"author_id":  1,
				"status":     1,
			})
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
			_, err := GetProblemDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": pid},
				bson.M{"$inc": bson.M{
					"accept": acceptDelta,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", pid)
			}
		}

		// 4. 批量更新 UserId 表的 accept 计数
		for uid, acceptDelta := range userAcceptDelta {
			_, err := GetUserDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": uid},
				bson.M{"$inc": bson.M{
					"accept": acceptDelta,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", uid)
			}
		}

		return nil, nil
	})
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

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. 查找最近提交
		findFilter := bson.D{
			{"origin_oj", bson.M{"$exists": false}},
			{"$or", bson.A{
				bson.M{"origin_oj": ""},
				bson.M{"origin_oj": nil},
			}},
		}
		findOpts := options.Find().
			SetSort(bson.D{{"_id", -1}}).
			SetProjection(bson.M{
				"_id":        1,
				"problem_id": 1,
				"author_id":  1,
				"status":     1,
			}).
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
			_, err := GetProblemDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": pid},
				bson.M{"$inc": bson.M{
					"accept": acceptDelta,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update problem accept count for problem %s", pid)
			}
		}

		// 4. 批量更新 UserId 表的 accept 计数
		for uid, acceptDelta := range userAcceptDelta {
			_, err := GetUserDao().collection.UpdateOne(sessCtx,
				bson.M{"_id": uid},
				bson.M{"$inc": bson.M{
					"accept": acceptDelta,
				}},
			)
			if err != nil {
				return nil, metaerror.Wrap(err, "failed to update user accept count for user %d", uid)
			}
		}

		return nil, nil
	})
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

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
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
			SetProjection(bson.M{
				"_id":        1,
				"problem_id": 1,
				"author_id":  1,
				"status":     1,
			})

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
				models = append(models, mongo.NewUpdateOneModel().
					SetFilter(bson.M{"_id": pid}).
					SetUpdate(bson.M{"$inc": bson.M{"accept": delta}}))
			}
			if _, err := GetProblemDao().collection.BulkWrite(sessCtx, models); err != nil {
				return nil, metaerror.Wrap(err, "update problem fail")
			}
		}

		// 更新 User 表
		if len(userAcceptDelta) > 0 {
			var models []mongo.WriteModel
			for uid, delta := range userAcceptDelta {
				models = append(models, mongo.NewUpdateOneModel().
					SetFilter(bson.M{"_id": uid}).
					SetUpdate(bson.M{"$inc": bson.M{"accept": delta}}))
			}
			if _, err := GetUserDao().collection.BulkWrite(sessCtx, models); err != nil {
				return nil, metaerror.Wrap(err, "update user fail")
			}
		}

		// 成功处理一页
		lastID = lastDocID
		slog.Info("RejudgePage success", "count", len(ids), "lastID", lastID)
		return nil, nil
	})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return -1, nil
	}
	if err != nil {
		return lastID, metaerror.Wrap(err, "transaction failed")
	}

	return lastID, nil
}
