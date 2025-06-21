package service

import (
	"context"
	foundationcontest "foundation/foundation-contest"
	foundationservice "foundation/foundation-service"
	"log/slog"
	metamysql "meta/meta-mysql"
	migratedao "migrate/dao"
	"slices"
	"strconv"
	"time"

	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	"meta/singleton"
)

type MigrateContestEojService struct {
}

var singletonMigrateContestEojService = singleton.Singleton[MigrateContestEojService]{}

func GetMigrateContestEojService() *MigrateContestEojService {
	return singletonMigrateContestEojService.GetInstance(
		func() *MigrateContestEojService {
			s := &MigrateContestEojService{}
			return s
		},
	)
}

type EojContest struct {
	Id          int        `gorm:"column:id"`
	Title       string     `gorm:"column:title"`
	Description string     `gorm:"column:description"`
	ContestType int        `gorm:"column:contest_type"`
	StartTime   *time.Time `gorm:"column:start_time"`
	EndTime     *time.Time `gorm:"column:end_time"`
	CreateTime  time.Time  `gorm:"column:create_time"`
}

func (EojContest) TableName() string {
	return "contest_contest"
}

type EojContestAuthor struct {
	Id        int `gorm:"column:id"`
	ContestId int `gorm:"column:contest_id"`
	UserId    int `gorm:"column:user_id"`
}

func (EojContestAuthor) TableName() string {
	return "contest_contest_authors"
}

type EojContestManager struct {
	Id        int `gorm:"column:id"`
	ContestId int `gorm:"column:contest_id"`
	UserId    int `gorm:"column:user_id"`
}

func (EojContestManager) TableName() string {
	return "contest_contest_managers"
}

type EojContestVolunteer struct {
	Id        int `gorm:"column:id"`
	ContestId int `gorm:"column:contest_id"`
	UserId    int `gorm:"column:user_id"`
}

func (EojContestVolunteer) TableName() string {
	return "contest_contest_volunteers"
}

type EojContestProblem struct {
	Id         int    `gorm:"column:id"`
	Identifier string `gorm:"column:identifier"`
	Weight     int    `gorm:"column:weight"`
	ContestId  int    `gorm:"column:contest_id"`
	ProblemId  int    `gorm:"column:problem_id"`
}

func (EojContestProblem) TableName() string {
	return "contest_contestproblem"
}

type EojContestSubmission struct {
	ContestId int `gorm:"column:contest_id"`
	AuthorId  int `gorm:"column:author_id"`
}

func (EojContestSubmission) TableName() string {
	return "submission_submission"
}

func (s *MigrateContestEojService) Start() error {
	ctx := context.Background()

	slog.Info("migrate contest start")

	ignoreContest := []int{
		4, 5, 8, 14, 44, 46,
	}

	// 初始化 GORM 客户端
	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var contestModels []EojContest
	if err := eojDb.
		Model(&EojContest{}).
		Find(&contestModels).Error; err != nil {
		return metaerror.Wrap(err, "query eoj contest failed")
	}

	for _, contestModel := range contestModels {
		if slices.Contains(ignoreContest, contestModel.Id) {
			slog.Info("ignore eoj contest", "id", contestModel.Id, "title", contestModel.Title)
			continue
		}

		var contestModelManagers []EojContestManager
		if err := eojDb.
			Model(&EojContestManager{}).
			Where("contest_id = ?", contestModel.Id).
			Find(&contestModelManagers).Error; err != nil {
			return metaerror.Wrap(err, "query managers failed")
		}
		var contestModelAuthors []EojContestAuthor
		if err := eojDb.
			Model(&EojContestAuthor{}).
			Where("contest_id = ?", contestModel.Id).
			Find(&contestModelAuthors).Error; err != nil {
			return metaerror.Wrap(err, "query authors failed")
		}
		var contestModelVolunteers []EojContestVolunteer
		if err := eojDb.
			Model(&EojContestVolunteer{}).
			Where("contest_id = ?", contestModel.Id).
			Find(&contestModelVolunteers).Error; err != nil {
			return metaerror.Wrap(err, "query volunteers failed")
		}

		userId := 3441

		var authMembers []int
		if len(contestModelManagers) > 0 {
			userId = contestModelManagers[0].UserId
			realId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(userId))
			if err != nil {
				return metaerror.Wrap(err, "get eoj user mark failed")
			}
			if realId == nil {
				return metaerror.Wrap(err, "get eoj user mark failed")
			}
			userId, err = strconv.Atoi(*realId)
			if err != nil {
				return metaerror.Wrap(err, "convert eoj user mark to int failed")
			}
			for _, manager := range contestModelManagers {
				realId, err = migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(manager.UserId))
				if err != nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				if realId == nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				authUserId, err := strconv.Atoi(*realId)
				if err != nil {
					return metaerror.Wrap(err, "convert eoj user mark to int failed")
				}
				authMembers = append(authMembers, authUserId)
			}
		}
		var authorMembers []int
		if len(contestModelAuthors) > 0 {
			for _, author := range contestModelAuthors {
				realId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(author.UserId))
				if err != nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				if realId == nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				authorUserId, err := strconv.Atoi(*realId)
				if err != nil {
					return metaerror.Wrap(err, "convert eoj user mark to int failed")
				}
				authorMembers = append(authorMembers, authorUserId)
			}
		}
		var volunteerMembers []int
		if len(contestModelVolunteers) > 0 {
			for _, volunteer := range contestModelVolunteers {
				realId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(volunteer.UserId))
				if err != nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				if realId == nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				authorUserId, err := strconv.Atoi(*realId)
				if err != nil {
					return metaerror.Wrap(err, "convert eoj user mark to int failed")
				}
				volunteerMembers = append(volunteerMembers, authorUserId)
			}
		}
		var contestModelProblems []EojContestProblem
		var realProblemIds []string
		if err := eojDb.
			Model(&EojContestProblem{}).
			Where("contest_id = ?", contestModel.Id).
			Find(&contestModelProblems).Error; err != nil {
			return metaerror.Wrap(err, "query problems failed")
		}

		var problems []*foundationmodel.ContestProblem
		for _, problemModel := range contestModelProblems {

			newId, err := migratedao.GetMigrateMarkDao().GetMark(
				ctx,
				"eoj-problem",
				strconv.Itoa(problemModel.ProblemId),
			)
			if err != nil {
				return err
			}
			realProblemIds = append(realProblemIds, *newId)

			problems = append(
				problems, foundationmodel.NewContestProblemBuilder().
					ProblemId(*newId).
					ViewId(nil).                                                              // 题目描述Id，默认为nil
					Weight(0).                                                                // 分数默认为0
					Index(foundationcontest.GetContestProblemIndex(problemModel.Identifier)). // 索引从1开始
					Build(),
			)
		}

		if contestModel.ContestType == 0 {

			contestService := foundationservice.GetContestService()
			contest := foundationmodel.NewContestBuilder().
				Title(contestModel.Title).
				Description(contestModel.Description).
				StartTime(*contestModel.StartTime).
				EndTime(*contestModel.EndTime).
				OwnerId(userId).
				CreateTime(contestModel.CreateTime).
				UpdateTime(contestModel.CreateTime).
				Problems(problems).
				Private(false).
				Build()

			newContestId, err := migratedao.GetMigrateMarkDao().GetMark(
				ctx,
				"eoj-contest",
				strconv.Itoa(contestModel.Id),
			)
			if err != nil {
				return err
			}
			var finalContestId string
			if newContestId != nil {
				newId, err := strconv.Atoi(*newContestId)
				if err != nil {
					return err
				}
				err = contestService.UpdateContest(ctx, newId, contest)
				if err != nil {
					return err
				}
				finalContestId = *newContestId
			} else {
				err := contestService.InsertContest(ctx, contest)
				if err != nil {
					return metaerror.Wrap(err, "insert contest failed")
				}
				finalContestId = strconv.Itoa(contest.Id)
			}
			err = migratedao.GetMigrateMarkDao().Mark(
				ctx,
				"eoj-contest",
				strconv.Itoa(contestModel.Id),
				finalContestId,
			)
			if err != nil {
				return metaerror.Wrap(err, "mark eoj contest failed")
			}
		} else {
			var authorIds []int
			if err := eojDb.
				Model(&EojContestSubmission{}).
				Select("DISTINCT author_id").
				Where("contest_id = ?", contestModel.Id).
				Pluck("author_id", &authorIds).Error; err != nil {
				return metaerror.Wrap(err, "query distinct author ids failed")
			}
			var realAuthorIds []int
			for _, authorId := range authorIds {
				realId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(authorId))
				if err != nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				if realId == nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				realAuthorId, err := strconv.Atoi(*realId)
				if err != nil {
					return metaerror.Wrap(err, "convert eoj user mark to int failed")
				}
				realAuthorIds = append(realAuthorIds, realAuthorId)
			}

			collection := foundationmodel.NewCollectionBuilder().
				Title(contestModel.Title).
				Description(contestModel.Description).
				StartTime(contestModel.StartTime).
				EndTime(contestModel.EndTime).
				OwnerId(userId).
				Problems(realProblemIds).
				Members(realAuthorIds).
				CreateTime(contestModel.CreateTime).
				UpdateTime(contestModel.CreateTime).
				Build()

			collectionService := foundationservice.GetCollectionService()

			newCollectionId, err := migratedao.GetMigrateMarkDao().GetMark(
				ctx,
				"eoj-collection",
				strconv.Itoa(contestModel.Id),
			)
			if err != nil {
				return metaerror.Wrap(err, "get eoj collection mark failed")
			}
			var finalContestId string
			if newCollectionId != nil {
				newId, err := strconv.Atoi(*newCollectionId)
				if err != nil {
					return metaerror.Wrap(err, "convert eoj collection mark to int failed")
				}
				err = collectionService.UpdateCollection(ctx, newId, collection)
				if err != nil {
					return metaerror.Wrap(err, "update collection failed")
				}
				finalContestId = *newCollectionId
			} else {
				err = collectionService.InsertCollection(ctx, collection)
				if err != nil {
					return metaerror.Wrap(err, "insert collection failed")
				}
				finalContestId = strconv.Itoa(collection.Id)
			}
			err = migratedao.GetMigrateMarkDao().Mark(
				ctx,
				"eoj-collection",
				strconv.Itoa(contestModel.Id),
				finalContestId,
			)
			if err != nil {
				return metaerror.Wrap(err, "mark eoj collection failed")
			}
		}
	}

	return nil
}
