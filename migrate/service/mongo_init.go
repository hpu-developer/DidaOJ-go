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
	return nil
}
