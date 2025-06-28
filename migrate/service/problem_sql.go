package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	metastring "meta/meta-string"
	"meta/singleton"
)

type MigrateProblemSqlService struct {
}

var singletonMigrateProblemSqlService = singleton.Singleton[MigrateProblemSqlService]{}

func GetMigrateProblemSqlService() *MigrateProblemSqlService {
	return singletonMigrateProblemSqlService.GetInstance(
		func() *MigrateProblemSqlService {
			return &MigrateProblemSqlService{}
		},
	)
}

func (s *MigrateProblemSqlService) Start(ctx context.Context) error {

	tagList, _, err := foundationdaomongo.GetProblemTagDao().GetProblemTagList(ctx, -1)
	if err != nil {
		return err
	}
	slog.Info("tagList", "tagList", len(tagList))

	for _, tag := range tagList {
		err := foundationdao.GetTagDao().InsertTag(ctx, tag.Name)
		if err != nil {
			return err
		}
	}

	problemList, err := foundationdaomongo.GetProblemDao().GetProblemListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("problemList", "problemList", len(problemList))

	for _, problem := range problemList {
		if problem.Id == "VJUDGE-51Nod-1000" {
			continue
		}
		inserter := problem.CreatorId
		var originAuthor *string
		if problem.OriginOj == nil {
			if inserter <= 0 {
				inserter = 3
			}
		} else {
			if problem.CreatorNickname != nil {
				originAuthor = problem.CreatorNickname
			}
		}
		var sourcePtr *string
		if problem.Source != nil {
			finalSource := metastring.GetTextEllipsis(*problem.Source, 100)
			sourcePtr = &finalSource
		}

		newProblem := foundationmodel.NewProblemBuilder().
			Key(problem.Id).
			Title(problem.Title).
			Description(problem.Description).
			Source(sourcePtr).
			TimeLimit(problem.TimeLimit).
			MemoryLimit(problem.MemoryLimit).
			JudgeType(problem.JudgeType).
			Inserter(inserter).
			InsertTime(problem.InsertTime).
			Modifier(inserter).
			ModifyTime(problem.UpdateTime).
			Accept(problem.Accept).
			Attempt(problem.Attempt).
			Private(problem.Private).
			Build()

		if problem.OriginOj == nil {
			newProblemLocal := foundationmodel.NewProblemLocalBuilder().
				JudgeMd5(problem.JudgeMd5).
				Build()
			err := foundationdao.GetProblemDao().InsertProblemLocal(ctx, newProblem, newProblemLocal)
			if err != nil {
				return err
			}
		} else {
			newProblemRemote := foundationmodel.NewProblemRemoteBuilder().
				OriginOj(*problem.OriginOj).
				OriginId(*problem.OriginId).
				OriginUrl(*problem.OriginUrl).
				OriginAuthor(originAuthor).
				Build()
			err := foundationdao.GetProblemDao().InsertProblemRemote(ctx, newProblem, newProblemRemote)
			if err != nil {
				return err
			}
		}

		err := foundationdao.GetProblemTagDao().UpdateProblemTags(ctx, newProblem.Id, problem.Tags)
		if err != nil {
			return err
		}

		err = foundationdao.GetProblemMemberAuthDao().UpdateProblemMemberAuths(ctx, newProblem.Id, problem.AuthMembers)
		if err != nil {
			return err
		}

	}

	return nil
}
