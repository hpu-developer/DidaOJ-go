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

type DiscussDao struct {
	collection *mongo.Collection
}

var singletonDiscussDao = singleton.Singleton[DiscussDao]{}

func GetDiscussDao() *DiscussDao {
	return singletonDiscussDao.GetInstance(
		func() *DiscussDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var DiscussDao DiscussDao
			DiscussDao.collection = client.
				Database("didaoj").
				Collection("discuss_migrate")
			return &DiscussDao
		},
	)
}

func (d *DiscussDao) GetDiscussList(ctx context.Context) ([]*foundationmodel.Discuss, error) {
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
	var list []*foundationmodel.Discuss
	if err = cursor.All(ctx, &list); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode documents")
	}
	return list, nil
}
