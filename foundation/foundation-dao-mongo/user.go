package foundationdaomongo

import (
	"context"
	"errors"
	foundationmodel "foundation/foundation-model-mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	metaerror "meta/meta-error"
	metamongo "meta/meta-mongo"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"strings"
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
		Keys: bson.D{{Key: "username_lower", Value: 1}}, // 1表示升序索引
		Options: options.Index().
			SetUnique(true),
	}
	_, err := d.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	return nil
}

func (d *UserDao) GetUserListAll(ctx context.Context) ([]*foundationmodel.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := d.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, metaerror.Wrap(err, "find all users error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err, "close cursor error")
		}
	}(cursor, ctx)
	var users []*foundationmodel.User
	for cursor.Next(ctx) {
		var user foundationmodel.User
		if err := cursor.Decode(&user); err != nil {
			return nil, metaerror.Wrap(err, "decode user error")
		}
		users = append(users, &user)
	}
	return users, nil
}

func (d *UserDao) InsertUser(ctx context.Context, user *foundationmodel.User) error {
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
			seq, err := GetCounterDao().GetNextSequence(sc, "user_id")
			if err != nil {
				return nil, err
			}
			// 更新 user 的 ID
			user.Id = seq
			// 插入新的 UserId
			_, err = d.collection.InsertOne(sc, user)
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

func (d *UserDao) GetUser(ctx context.Context, userId int) (*foundationmodel.User, error) {
	filter := bson.M{
		"_id": userId,
	}
	var User foundationmodel.User
	if err := d.collection.FindOne(ctx, filter).Decode(&User); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find UserId error")
	}
	return &User, nil
}

func (d *UserDao) GetUserByUsername(ctx context.Context, username string) (*foundationmodel.User, error) {
	filter := bson.M{
		"username_lower": strings.ToLower(username),
	}
	var User foundationmodel.User
	if err := d.collection.FindOne(ctx, filter).Decode(&User); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find UserId error")
	}
	return &User, nil
}

func (d *UserDao) GetInfo(ctx *gin.Context, username string) (*foundationmodel.UserInfo, error) {
	filter := bson.M{
		"username_lower": strings.ToLower(username),
	}
	opts := options.FindOne().SetProjection(
		bson.M{
			"_id":           1,
			"username":      1,
			"nickname":      1,
			"email":         1,
			"qq":            1,
			"slogan":        1,
			"organization":  1,
			"reg_time":      1,
			"accept":        1,
			"attempt":       1,
			"checkin_count": 1,
			"vjudge_id":     1,
		},
	)
	var User foundationmodel.UserInfo
	if err := d.collection.FindOne(ctx, filter, opts).Decode(&User); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find UserId error")
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

func (d *UserDao) GetUserAccountInfos(
	ctx context.Context,
	userIds []int,
) ([]*foundationmodel.UserAccountInfo, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": userIds,
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

func (d *UserDao) GetUserAccountInfoByUsernames(
	ctx context.Context,
	usernames []string,
) ([]*foundationmodel.UserAccountInfo, error) {
	for i, username := range usernames {
		usernames[i] = strings.ToLower(username)
	}
	filter := bson.M{
		"username_lower": bson.M{
			"$in": usernames,
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

func (d *UserDao) GetUsersRankInfo(ctx context.Context, userId []int) ([]*foundationmodel.UserRankInfo, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": userId,
		},
	}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1, "username": 1, "nickname": 1, "slogan": 1})
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
	var result []*foundationmodel.UserRankInfo
	for cursor.Next(ctx) {
		var user foundationmodel.UserRankInfo
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
	findOptions := options.FindOne().SetProjection(
		bson.M{
			"_id":      1,
			"username": 1,
			"nickname": 1,
			"password": 1,
			"roles":    1,
		},
	)
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
		"username_lower": strings.ToLower(username),
	}
	findOptions := options.FindOne().SetProjection(
		bson.M{
			"_id":      1,
			"username": 1,
			"nickname": 1,
			"password": 1,
			"roles":    1,
		},
	)
	var result foundationmodel.UserLogin
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find user password error")
	}
	return &result, nil
}

func (d *UserDao) GetUserIdByUsername(ctx context.Context, username string) (int, error) {
	filter := bson.M{
		"username_lower": strings.ToLower(username),
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

func (d *UserDao) GetEmailByUsername(ctx context.Context, username string) (*string, error) {
	filter := bson.M{
		"username_lower": strings.ToLower(username),
	}
	findOptions := options.FindOne().SetProjection(bson.M{"_id": 1, "email": 1})
	var result struct {
		Email *string `bson:"email"`
	}
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find user email error")
	}
	return result.Email, nil
}

func (d *UserDao) GetUserRoles(ctx context.Context, userId int) ([]string, error) {
	filter := bson.M{
		"_id": userId,
	}
	findOptions := options.FindOne().SetProjection(bson.M{"_id": 1, "roles": 1})
	var result struct {
		Roles []string `bson:"roles"`
	}
	if err := d.collection.FindOne(ctx, filter, findOptions).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find user roles error")
	}
	return result.Roles, nil
}

func (d *UserDao) GetUserIds(ctx *gin.Context, usernames []string) ([]int, error) {
	for i, username := range usernames {
		usernames[i] = strings.ToLower(username)
	}
	filter := bson.M{
		"username_lower": bson.M{
			"$in": usernames,
		},
	}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, metaerror.Wrap(err, "find user ids error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(err, "close cursor error")
		}
	}(cursor, ctx)
	var result []int
	for cursor.Next(ctx) {
		var user struct {
			Id int `bson:"_id"`
		}
		if err := cursor.Decode(&user); err != nil {
			return nil, metaerror.Wrap(err, "decode user id error")
		}
		result = append(result, user.Id)
	}
	return result, nil
}

func (d *UserDao) UpdateUser(ctx context.Context, userId int, User *foundationmodel.User) error {
	filter := bson.D{
		{"_id", userId},
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

func (d *UserDao) UpdatePassword(ctx *gin.Context, username string, encode string) error {
	filter := bson.M{
		"username_lower": strings.ToLower(username),
	}
	update := bson.M{
		"$set": bson.M{
			"password": encode,
		},
	}
	updateOptions := options.Update()
	_, err := d.collection.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return metaerror.Wrap(err, "failed to update password")
	}
	return nil
}

func (d *UserDao) GetRankAcAll(ctx *gin.Context, page int, pageSize int) ([]*foundationmodel.UserRank, int, error) {
	filter := bson.M{
		"accept": bson.M{
			"$gt": 0,
		},
	}
	limit := int64(pageSize)
	skip := int64((page - 1) * pageSize)

	opts := options.Find().
		SetProjection(
			bson.M{
				"_id":      1,
				"username": 1,
				"nickname": 1,
				"slogan":   1,
				"accept":   1,
				"attempt":  1,
			},
		).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(
			bson.D{
				{Key: "accept", Value: -1},
				{Key: "attempt", Value: 1},
				{Key: "_id", Value: 1},
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
	var list []*foundationmodel.UserRank
	if err = cursor.All(ctx, &list); err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to decode documents, page: %d", page)
	}
	return list, int(totalCount), nil
}

func (d *UserDao) FilterValidUserIds(ctx *gin.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	findOptions := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := d.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, metaerror.Wrap(err, "find valid user ids error")
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close cursor error"))
		}
	}(cursor, ctx)
	var result []int
	for cursor.Next(ctx) {
		var user struct {
			Id int `bson:"_id"`
		}
		if err := cursor.Decode(&user); err != nil {
			return nil, metaerror.Wrap(err, "decode user id error")
		}
		result = append(result, user.Id)
	}
	return result, nil
}
