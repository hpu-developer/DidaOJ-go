package foundationdao

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
)

type UserDao struct {
	collection *mongo.Collection
}

var singletonUserDao = singleton.Singleton[UserDao]{}

func GetUserDao() *UserDao {
	return singletonUserDao.GetInstance(
		func() *UserDao {
			mongoSubsystem := metamongo.GetSubsystem()
			if mongoSubsystem == nil {
				return nil
			}
			client := mongoSubsystem.GetClient()
			var UserDao UserDao
			UserDao.collection = client.
				Database("didaoj").
				Collection("user")
			return &UserDao
		},
	)
}

func (d *UserDao) InitDao(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}}, // 1表示升序索引
		Options: options.Index().SetUnique(true),
	}
	_, err := d.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return nil
}

func (d *UserDao) GetUser(ctx context.Context, userId int) (*foundationmodel.User, error) {
	filter := bson.M{
		"_id": userId,
	}
	var User foundationmodel.User
	if err := d.collection.FindOne(ctx, filter).Decode(&User); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find User error")
	}
	return &User, nil
}

func (d *UserDao) GetUserAccountInfo(ctx context.Context, userId int) (*foundationmodel.UserAccountInfo, error) {
	filter := bson.M{
		"_id": userId,
	}
	findOptions := options.FindOne().SetProjection(bson.M{"_id": 1, "username": 1, "nickname": 1})
	var result foundationmodel.UserAccountInfo
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find user account info error")
	}
	return &result, nil
}

func (d *UserDao) GetUsersAccountInfo(ctx context.Context, userId []int) ([]*foundationmodel.UserAccountInfo, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": userId,
		},
	}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1, "username": 1, "nickname": 1})
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
	var result []*foundationmodel.UserAccountInfo
	for cursor.Next(ctx) {
		var user foundationmodel.UserAccountInfo
		if err := cursor.Decode(&user); err != nil {
			return nil, metaerror.Wrap(err, "decode user account info error")
		}
		result = append(result, &user)
	}
	return result, nil
}

func (d *UserDao) GetUserLogin(ctx context.Context, userId int) (*foundationmodel.UserLogin, error) {
	filter := bson.M{
		"_id": userId,
	}
	findOptions := options.FindOne().SetProjection(bson.M{"_id": 1, "username": 1, "nickname": 1, "password": 1})
	var result foundationmodel.UserLogin
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find user password error")
	}
	return &result, nil
}

func (d *UserDao) GetUserLoginByUsername(ctx context.Context, username string) (*foundationmodel.UserLogin, error) {
	filter := bson.M{
		"username": username,
	}
	findOptions := options.FindOne().SetProjection(bson.M{"_id": 1, "username": 1, "nickname": 1, "password": 1})
	var result foundationmodel.UserLogin
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find user password error")
	}
	return &result, nil
}

func (d *UserDao) UpdateUser(ctx context.Context, key string, User *foundationmodel.User) error {
	filter := bson.D{
		{"_id", key},
	}
	update := bson.M{
		"$set": User,
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to save tapd subscription")
	}
	return nil
}

func (d *UserDao) GetUserIdByUsername(ctx context.Context, username string) (int, error) {
	filter := bson.M{
		"username": username,
	}
	findOptions := options.FindOne().SetProjection(bson.M{"_id": 1})
	var result struct {
		UserId int `bson:"_id"`
	}
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		return 0, metaerror.Wrap(err, "find user id by username error, %s", username)
	}
	return result.UserId, nil
}

func (d *UserDao) UpdateUsers(ctx context.Context, users []*foundationmodel.User) error {
	for _, user := range users {
		filter := bson.D{
			{"_id", user.Id},
		}
		update := bson.M{
			"$set": user,
		}
		updateOptions := options.Update().SetUpsert(true)
		_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
		if err != nil {
			return metaerror.Wrap(err, "failed to update user")
		}
	}
	return nil
}
