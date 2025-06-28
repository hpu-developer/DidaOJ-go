package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
)

type MigrateDiscussCommentSqlService struct {
}

var singletonMigrateDiscussCommentSqlService = singleton.Singleton[MigrateDiscussCommentSqlService]{}

func GetMigrateDiscussCommentSqlService() *MigrateDiscussCommentSqlService {
	return singletonMigrateDiscussCommentSqlService.GetInstance(
		func() *MigrateDiscussCommentSqlService {
			return &MigrateDiscussCommentSqlService{}
		},
	)
}

func (s *MigrateDiscussCommentSqlService) Start(ctx context.Context) error {

	discussCommentList, err := foundationdaomongo.GetDiscussCommentDao().GetListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("discussCommentList", "discussCommentList", len(discussCommentList))

	for _, discussComment := range discussCommentList {

		newDiscussComment := foundationmodel.NewDiscussCommentBuilder().
			Id(discussComment.Id).
			DiscussId(discussComment.DiscussId).
			Content(discussComment.Content).
			Banned(false).
			Inserter(discussComment.AuthorId).
			InsertTime(discussComment.InsertTime).
			Modifier(discussComment.AuthorId).
			ModifyTime(discussComment.UpdateTime).
			Build()

		err = foundationdao.GetDiscussCommentDao().InsertDiscussComment(ctx, newDiscussComment)
		if err != nil {
			return err
		}
	}

	return nil
}
