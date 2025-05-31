package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
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

func (s *CollectionService) GetCollection(ctx context.Context, id int) (*foundationmodel.Collection, error) {
	collection, err := foundationdao.GetCollectionDao().GetCollection(ctx, id)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, nil
	}
	collectionProblems := map[string]*foundationmodel.CollectionProblem{}
	for _, problem := range collection.Problems {
		collectionProblems[problem] = &foundationmodel.CollectionProblem{
			ProblemId: problem,
		}
	}
	var problemIds []string
	for _, problem := range collection.Problems {
		problemIds = append(problemIds, problem)
	}
	problems, err := foundationdao.GetProblemDao().GetProblemListTitle(ctx, problemIds)
	if err != nil {
		return nil, err
	}
	for _, problem := range problems {
		if collectionProblem, ok := collectionProblems[problem.Id]; ok {
			collectionProblem.Title = &problem.Title
		}
	}
	judgeAccepts, err := foundationdao.GetJudgeJobDao().GetProblemTimeViewAttempt(
		ctx,
		collection.StartTime,
		collection.EndTime,
		problemIds,
	)
	if err != nil {
		return nil, err
	}
	for _, judgeAccept := range judgeAccepts {
		if collectionProblem, ok := collectionProblems[judgeAccept.Id]; ok {
			collectionProblem.Accept = judgeAccept.Accept
			collectionProblem.Attempt = judgeAccept.Attempt
		}
	}
	return collection, err
}

func (s *CollectionService) GetCollectionList(
	ctx context.Context,
	page int,
	pageSize int,
) ([]*foundationmodel.Collection, int, error) {
	collections, totalCount, err := foundationdao.GetCollectionDao().GetCollectionList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var userIds []int
	for _, collection := range collections {
		userIds = append(userIds, collection.OwnerId)
	}
	users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
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
	return collections, totalCount, nil
}

func (s *CollectionService) InsertCollection(ctx context.Context, collection *foundationmodel.Collection) error {
	return foundationdao.GetCollectionDao().InsertCollection(ctx, collection)
}

func (s *CollectionService) GetCollectionRanks(ctx context.Context, id int) (
	*time.Time,
	*time.Time,
	[]string,
	[]*foundationmodel.CollectionRank,
	error,
) {
	collectionView, err := foundationdao.GetCollectionDao().GetCollectionRankView(ctx, id)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	problemMap := make(map[string]string)
	for _, problem := range collectionView.Problems {
		problemMap[problem] = problem
	}
	collectionRanks, err := foundationdao.GetJudgeJobDao().GetCollectionRanks(
		ctx,
		collectionView.StartTime,
		collectionView.EndTime,
		collectionView.Problems,
		collectionView.Members,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if len(collectionRanks) > 0 {
		var userIds []int
		for _, collectionRank := range collectionRanks {
			userIds = append(userIds, collectionRank.AuthorId)
		}
		users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, collectionRank := range collectionRanks {
			if user, ok := userMap[collectionRank.AuthorId]; ok {
				collectionRank.AuthorUsername = &user.Username
				collectionRank.AuthorNickname = &user.Nickname
			}
		}
	}
	return collectionView.StartTime, collectionView.EndTime, collectionView.Problems, collectionRanks, nil
}
