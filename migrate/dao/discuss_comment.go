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

type DiscussCommentDao struct {
	collection *mongo.Collection
}

var singletonDiscussCommentDao = singleton.Singleton[DiscussCommentDao]{}

func GetDiscussCommentDao() *DiscussCommentDao {
	return singletonDiscussCommentDao.GetInstance(
		func() *DiscussCommentDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var DiscussCommentDao DiscussCommentDao
			DiscussCommentDao.collection = client.
				Database("didaoj").
				Collection("discuss_comment_migrate")
			return &DiscussCommentDao
		},
	)
}

func (d *DiscussCommentDao) GetDiscussCommentList(ctx context.Context) ([]*foundationmodel.DiscussComment, error) {
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
	var list []*foundationmodel.DiscussComment
	if err = cursor.All(ctx, &list); err != nil {
		return nil, metaerror.Wrap(err, "failed to decode documents")
	}
	return list, nil
}
