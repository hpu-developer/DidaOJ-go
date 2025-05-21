package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	"meta/singleton"
)

type MongoInitService struct{}

var singletonMongoInitService = singleton.Singleton[MongoInitService]{}

func GetMongoInitService() *MongoInitService {
	return singletonMongoInitService.GetInstance(
		func() *MongoInitService {
			return &MongoInitService{}
		},
	)
}

func (s *MongoInitService) Start() error {
	ctx := context.Background()
	var err error
	err = foundationdao.GetCounterDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetProblemDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetProblemTagDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetUserDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetContestDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetDiscussDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetDiscussCommentDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetDiscussTagDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetJudgerDao().InitDao(ctx)
	if err != nil {
		return nil
	}
	err = foundationdao.GetJudgeJobDao().InitDao(ctx)
	if err != nil {
		return nil
	}

	return nil
}
