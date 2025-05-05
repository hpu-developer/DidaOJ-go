package foundationdao

import (
	"context"
	"errors"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
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
		seq, err := GetCounterDao().GetNextSequence(sc, "jduge_id")
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

func (d *JudgeJobDao) GetJudgeJobList(ctx context.Context,
	page int,
	pageSize int,
) ([]*foundationmodel.JudgeJob,
	int,
	error,
) {
	filter := bson.M{}
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
			"author":       1,
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

func (d *JudgeJobDao) GetProblemAttemptStatus(ctx context.Context, problemIds []string, author int,
) (map[string]foundationmodel.ProblemAttemptStatus, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"author":     author,
			"problem_id": bson.M{"$in": problemIds},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$problem_id",
			"statusSum": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$status", 4}},
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

// GetJudgeJobListPendingJudge 获取待评测的 JudgeJob 列表，优先取最小的
func (d *JudgeJobDao) GetJudgeJobListPendingJudge(ctx context.Context, maxCount int) ([]*foundationmodel.JudgeJob, error) {
	filter := bson.M{
		"status": bson.M{
			"$in": []foundationjudge.JudgeStatus{foundationjudge.JudgeStatusInit, foundationjudge.JudgeStatusRejudge},
		},
	}
	// 按照 id 升序排列
	findOptions := options.Find().SetSort(bson.M{"_id": 1})
	if maxCount > 0 {
		findOptions.SetLimit(int64(maxCount))
	}
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, metaerror.Wrap(err, "find JudgeJob error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(cursor, ctx)
	var judgeSourceList []*foundationmodel.JudgeJob
	if err = cursor.All(ctx, &judgeSourceList); err != nil {
		return nil, metaerror.Wrap(err, "decode JudgeJob error")
	}
	return judgeSourceList, nil
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
			"judger":     judger,
			"status":     foundationjudge.JudgeStatusCompiling,
			"judge_time": metatime.GetTimeNow(),
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
	score int,
	time int,
	memory int,
) error {
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
