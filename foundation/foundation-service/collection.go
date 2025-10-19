package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationenum "foundation/foundation-enum"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"meta/singleton"
	"time"

	"github.com/gin-gonic/gin"
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
	collection *foundationview.CollectionDetail,
	joined bool,
	collectionProblems []*foundationview.ProblemViewList,
	tags []*foundationmodel.Tag,
	userAttempts map[int]foundationenum.ProblemAttemptStatus,
	err error,
) {
	collection, err = foundationdao.GetCollectionDao().GetCollectionDetail(ctx, id)
	if err != nil {
		return
	}
	if collection == nil {
		return
	}
	collection.Password = nil // 不真正返回密码
	if userId > 0 {
		joined, err = foundationdao.GetCollectionDao().CheckUserJoin(ctx, id, userId)
		if err != nil {
			return
		}
	}
	var collectionProblemIds []int
	collectionProblemIds, err = foundationdao.GetCollectionProblemDao().GetCollectionProblems(ctx, id)
	if err != nil {
		return
	}
	if len(collectionProblemIds) == 0 {
		return
	}
	collectionProblems, err = foundationdao.GetProblemDao().SelectProblemViewList(ctx, collectionProblemIds, false)
	if err != nil {
		return
	}
	var problemTags map[int][]int
	problemTags, err = foundationdao.GetProblemTagDao().GetProblemTagMap(ctx, collectionProblemIds)
	problemMap := make(map[int]*foundationview.ProblemViewList, len(collectionProblemIds))
	for _, problem := range collectionProblems {
		problemMap[problem.Id] = problem
	}
	var problemTagIds []int
	for problemId, tag := range problemTags {
		problemTagIds = append(problemTagIds, tag...)
		problemMap[problemId].Tags = tag
	}
	tags, err = foundationdao.GetTagDao().GetTags(ctx, problemTagIds)
	if err != nil {
		return
	}
	if userId > 0 {
		userAttempts, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx,
			userId,
			collectionProblemIds,
			-1,
			collection.StartTime,
			collection.EndTime,
		)
		if err != nil {
			return
		}
	}
	var problemAttempts []*foundationview.ProblemAttemptInfo
	problemAttempts, err = foundationdao.GetCollectionDao().GetProblemAttemptInfo(
		ctx,
		id,
		collectionProblemIds,
		collection.StartTime,
		collection.EndTime,
	)
	if err != nil {
		return
	}
	if len(problemAttempts) > 0 {
		for _, attempt := range problemAttempts {
			if problem, ok := problemMap[attempt.Id]; ok {
				problem.Accept = attempt.Accept
				problem.Attempt = attempt.Attempt
			}
		}
	}
	return
}

func (s *CollectionService) GetCollectionEdit(ctx context.Context, id int) (
	*foundationview.CollectionDetail,
	error,
) {
	collection, err := foundationdao.GetCollectionDao().GetCollectionDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, nil
	}
	collection.Problems, err = foundationdao.GetCollectionProblemDao().GetCollectionProblems(ctx, id)
	if err != nil {
		return nil, err
	}
	collection.Members, err = foundationdao.GetCollectionDao().GetCollectionMemberIds(ctx, id)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (s *CollectionService) GetCollectionList(
	ctx context.Context,
	page int,
	pageSize int,
) ([]*foundationview.CollectionList, int, error) {
	return foundationdao.GetCollectionDao().GetCollectionList(ctx, page, pageSize)
}

func (s *CollectionService) GetCollectionRanks(ctx context.Context, id int) (
	startTime *time.Time,
	endTime *time.Time,
	problems int,
	ranks []*foundationview.CollectionRank,
	err error,
) {
	var collection *foundationview.CollectionRankDetail
	collection, err = foundationdao.GetCollectionDao().GetCollectionRankDetail(ctx, id)
	if err != nil {
		return
	}
	if collection == nil {
		return
	}
	collection.Problems, err = foundationdao.GetCollectionProblemDao().GetCollectionProblems(ctx, id)
	if err != nil {
		return
	}
	ranks, err = foundationdao.GetCollectionDao().GetCollectionRank(ctx, id, collection)
	if err != nil {
		return
	}
	problems = len(collection.Problems)
	return
}

func (s *CollectionService) PostJoin(ctx *gin.Context, collectionId int, userId int) error {
	return foundationdaomongo.GetCollectionDao().PostJoin(ctx, collectionId, userId)
}

func (s *CollectionService) PostQuit(ctx *gin.Context, collectionId int, userId int) error {
	return foundationdaomongo.GetCollectionDao().PostQuit(ctx, collectionId, userId)
}

func (s *CollectionService) InsertCollection(
	ctx context.Context,
	collection *foundationmodel.Collection,
	problemIds []int,
	members []int,
) error {
	return foundationdao.GetCollectionDao().InsertCollection(ctx, collection, problemIds, members)
}

func (s *CollectionService) UpdateCollection(
	ctx context.Context,
	collection *foundationmodel.Collection,
	problemIds []int,
	members []int,
) error {
	return foundationdao.GetCollectionDao().UpdateCollection(ctx, collection, problemIds, members)
}
