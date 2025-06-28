package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
	"slices"
)

type MigrateCollectionSqlService struct {
}

var singletonMigrateCollectionSqlService = singleton.Singleton[MigrateCollectionSqlService]{}

func GetMigrateCollectionSqlService() *MigrateCollectionSqlService {
	return singletonMigrateCollectionSqlService.GetInstance(
		func() *MigrateCollectionSqlService {
			return &MigrateCollectionSqlService{}
		},
	)
}

func (s *MigrateCollectionSqlService) Start(ctx context.Context) error {

	collectionList, err := foundationdaomongo.GetCollectionDao().GetListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("collectionList", "collectionList", len(collectionList))

	for _, problemDaily := range collectionList {

		newCollection := foundationmodel.NewCollectionBuilder().
			Id(problemDaily.Id).
			Title(problemDaily.Title).
			Inserter(problemDaily.OwnerId).
			Modifier(problemDaily.OwnerId).
			InsertTime(problemDaily.CreateTime).
			ModifyTime(problemDaily.UpdateTime).
			Build()

		if problemDaily.Description != "" {
			newCollection.Description = &problemDaily.Description
		}

		var problemIds []int
		for _, problemKey := range problemDaily.Problems {
			problemId, err := foundationdao.GetProblemDao().GetProblemIdByKey(problemKey)
			if err != nil {
				return err
			}
			if slices.Contains(problemIds, problemId) {
				slog.Warn(
					"duplicate problem key in collection",
					"problemKey",
					problemKey,
					"collectionId",
					problemDaily.Id,
				)
				continue
			}
			problemIds = append(problemIds, problemId)
		}
		err = foundationdao.GetCollectionDao().InsertCollection(ctx, newCollection, problemIds, problemDaily.Members)
		if err != nil {
			return err
		}
	}

	return nil
}
