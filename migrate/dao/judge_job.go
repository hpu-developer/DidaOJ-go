package foundationdao

import (
	"context"
	"fmt"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
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
				Collection("judge_job_migrate")
			return &JudgeJobDao
		},
	)
}

func (d *JudgeJobDao) GetJudgeJobList(ctx context.Context) ([]*foundationmodel.JudgeJob, error) {
	filter := bson.M{}
	cursor, err := d.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find submissions: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "failed to close cursor"))
		}
	}(cursor, ctx)
	var list []*foundationmodel.JudgeJob
	if err = cursor.All(ctx, &list); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode documents")
	}
	return list, nil
}
