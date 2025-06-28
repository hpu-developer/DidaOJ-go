package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
	"github.com/gin-gonic/gin"
	"meta/singleton"
	"slices"
	"time"
)

type CollectionService struct {
}

var singletonCollectionService = singleton.Singleton[CollectionService]{}

func GetCollectionService() *CollectionService {
	return singletonCollectionService.GetInstance(
		func() *CollectionService {
			return &CollectionService{}
		},
	)
}

func (s *CollectionService) CheckEditAuth(ctx *gin.Context, collectionId int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageCollection)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		ownerId, err := foundationdaomongo.GetCollectionDao().GetCollectionOwnerId(ctx, collectionId)
		if err != nil {
			return userId, false, err
		}
		if ownerId != userId {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *CollectionService) CheckJoinAuth(ctx *gin.Context, collectionId int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageCollection)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		hasAuth, err = foundationdaomongo.GetCollectionDao().CheckJoinAuth(ctx, collectionId, userId)
		if err != nil {
			return userId, false, err
		}
	}
	return userId, hasAuth, nil
}

func (s *CollectionService) HasCollectionTitle(ctx *gin.Context, id int, title string) (bool, error) {
	return foundationdaomongo.GetCollectionDao().HasCollectionTitle(ctx, id, title)
}

func (s *CollectionService) GetCollection(ctx context.Context, id int, userId int) (
	*foundationmodel.Collection,
	[]*foundationmodel.CollectionProblem,
	bool,
	map[string]foundationmodel.ProblemAttemptStatus,
	error,
) {
	collection, err := foundationdaomongo.GetCollectionDao().GetCollection(ctx, id)
	if err != nil {
		return nil, nil, false, nil, err
	}
	if collection == nil {
		return nil, nil, false, nil, nil
	}
	ownerUser, err := foundationdaomongo.GetUserDao().GetUserAccountInfo(ctx, collection.OwnerId)
	if err != nil {
		return nil, nil, false, nil, err
	}
	collection.OwnerUsername = &ownerUser.Username
	collection.OwnerNickname = &ownerUser.Nickname

	joined := false
	var collectionProblemList []*foundationmodel.CollectionProblem
	var problemStatus map[string]foundationmodel.ProblemAttemptStatus
	if len(collection.Problems) > 0 {
		collectionProblems := map[string]*foundationmodel.CollectionProblem{}
		for _, problem := range collection.Problems {
			collectionProblems[problem] = &foundationmodel.CollectionProblem{
				Id: problem,
			}
		}
		var problemIds []string
		for _, problem := range collection.Problems {
			problemIds = append(problemIds, problem)
		}
		problems, err := foundationdaomongo.GetProblemDao().GetProblemListTitle(ctx, problemIds)
		if err != nil {
			return nil, nil, false, nil, err
		}
		for _, problem := range problems {
			if collectionProblem, ok := collectionProblems[problem.Id]; ok {
				collectionProblem.Title = &problem.Title
			}
		}
		if userId > 0 {
			problemStatus, err = foundationdaomongo.GetJudgeJobDao().GetProblemAttemptStatus(
				ctx,
				problemIds,
				userId,
				-1,
				collection.StartTime,
				collection.EndTime,
			)
			if err != nil {
				return nil, nil, false, nil, err
			}
		}

		var judgeAccepts []*foundationmodel.ProblemViewAttempt
		if len(collection.Members) > 0 {
			judgeAccepts, err = foundationdaomongo.GetJudgeJobDao().GetProblemTimeViewAttempt(
				ctx,
				collection.StartTime,
				collection.EndTime,
				problemIds,
				collection.Members,
			)
			if err != nil {
				return nil, nil, false, nil, err
			}
			for _, judgeAccept := range judgeAccepts {
				if collectionProblem, ok := collectionProblems[judgeAccept.Id]; ok {
					collectionProblem.Accept = judgeAccept.Accept
					collectionProblem.Attempt = judgeAccept.Attempt
				}
			}
			joined = slices.Contains(collection.Members, userId)
			// 不需要返回全部信息
			collection.Members = nil
		}
		for _, problem := range collection.Problems {
			collectionProblemList = append(collectionProblemList, collectionProblems[problem])
		}
	}
	return collection, collectionProblemList, joined, problemStatus, err
}

func (s *CollectionService) GetCollectionEdit(ctx context.Context, id int) (
	*foundationmodel.Collection,
	error,
) {
	collection, err := foundationdaomongo.GetCollectionDao().GetCollectionEdit(ctx, id)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, nil
	}
	ownerUser, err := foundationdaomongo.GetUserDao().GetUserAccountInfo(ctx, collection.OwnerId)
	if err != nil {
		return nil, err
	}
	collection.OwnerUsername = &ownerUser.Username
	collection.OwnerNickname = &ownerUser.Nickname
	return collection, err
}

func (s *CollectionService) GetCollectionList(
	ctx context.Context,
	page int,
	pageSize int,
) ([]*foundationmodel.Collection, int, error) {
	collections, totalCount, err := foundationdaomongo.GetCollectionDao().GetCollectionList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(collections) > 0 {
		var userIds []int
		for _, collection := range collections {
			userIds = append(userIds, collection.OwnerId)
		}
		users, err := foundationdaomongo.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, collection := range collections {
			if user, ok := userMap[collection.OwnerId]; ok {
				collection.OwnerUsername = &user.Username
				collection.OwnerNickname = &user.Nickname
			}
		}
	}
	return collections, totalCount, nil
}

func (s *CollectionService) InsertCollection(ctx context.Context, collection *foundationmodel.Collection) error {
	return foundationdaomongo.GetCollectionDao().InsertCollection(ctx, collection)
}

func (s *CollectionService) GetCollectionRanks(ctx context.Context, id int) (
	*time.Time,
	*time.Time,
	int,
	[]*foundationmodel.CollectionRank,
	error,
) {
	collectionView, err := foundationdaomongo.GetCollectionDao().GetCollectionRankView(ctx, id)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	var collectionRanks []*foundationmodel.CollectionRank
	if len(collectionView.Members) > 0 {
		users, err := foundationdaomongo.GetUserDao().GetUsersAccountInfo(ctx, collectionView.Members)
		if err != nil {
			return nil, nil, 0, nil, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, authorId := range collectionView.Members {
			if user, ok := userMap[authorId]; ok {
				collectionRanks = append(
					collectionRanks, &foundationmodel.CollectionRank{
						AuthorId:       authorId,
						AuthorUsername: &user.Username,
						AuthorNickname: &user.Nickname,
					},
				)
			}
		}
		if len(collectionView.Problems) > 0 {
			userAcMap, err := foundationdaomongo.GetJudgeJobDao().GetAcceptedProblemCount(
				ctx,
				collectionView.StartTime,
				collectionView.EndTime,
				collectionView.Problems,
				collectionView.Members,
			)
			if err != nil {
				return nil, nil, 0, nil, err
			}
			for _, collectionRank := range collectionRanks {
				if acCount, ok := userAcMap[collectionRank.AuthorId]; ok {
					collectionRank.Accept = acCount
				}
			}
		}
	}
	return collectionView.StartTime, collectionView.EndTime, len(collectionView.Problems), collectionRanks, nil
}

func (s *CollectionService) PostJoin(ctx *gin.Context, collectionId int, userId int) error {
	return foundationdaomongo.GetCollectionDao().PostJoin(ctx, collectionId, userId)
}

func (s *CollectionService) PostQuit(ctx *gin.Context, collectionId int, userId int) error {
	return foundationdaomongo.GetCollectionDao().PostQuit(ctx, collectionId, userId)
}

func (s *CollectionService) UpdateCollection(
	ctx context.Context,
	id int,
	collection *foundationmodel.Collection,
) error {
	return foundationdaomongo.GetCollectionDao().UpdateCollection(ctx, id, collection)
}
