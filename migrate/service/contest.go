package service

import (
	"context"
	"fmt"
	foundationenum "foundation/foundation-enum"
	"log/slog"
	metamysql "meta/meta-mysql"
	"sort"
	"strconv"
	"time"

	foundationdao "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
	metaerror "meta/meta-error"
	"meta/singleton"
)

type MigrateContestService struct {
	oldJolContestIdToNewContestId  map[int]int
	oldVhojContestIdToNewContestId map[int]int
}

var singletonMigrateContestService = singleton.Singleton[MigrateContestService]{}

func GetMigrateContestService() *MigrateContestService {
	return singletonMigrateContestService.GetInstance(
		func() *MigrateContestService {
			s := &MigrateContestService{}
			s.oldJolContestIdToNewContestId = make(map[int]int)
			s.oldVhojContestIdToNewContestId = make(map[int]int)
			return s
		},
	)
}

type JolContestProblem struct {
	ProblemID int `gorm:"column:problem_id"`
	ContestID int `gorm:"column:contest_id"`
	Num       int `gorm:"column:num"`
	Scores    int `gorm:"column:scores"`
}

func (JolContestProblem) TableName() string {
	return "contest_problem"
}

// JolContest GORM 模型定义
type JolContest struct {
	ContestID    int       `gorm:"column:contest_id"`
	Title        string    `gorm:"column:title"`
	StartTime    time.Time `gorm:"column:start_time"`
	EndTime      time.Time `gorm:"column:end_time"`
	Description  string    `gorm:"column:description"`
	Notification string    `gorm:"column:notification"`
	Langmask     int       `gorm:"column:langmask"`
	Password     *string   `gorm:"column:password"`
	Kind         int       `gorm:"column:kind"`
	UserId       string    `gorm:"column:user_id"`
}

func (JolContest) TableName() string {
	return "contest"
}

type VhojContest struct {
	ContestID      int       `gorm:"column:C_ID"`
	Title          string    `gorm:"column:C_TITLE"`
	Description    string    `gorm:"column:C_DESCRIPTION"`
	BeginTime      time.Time `gorm:"column:C_BEGINTIME"`
	EndTime        time.Time `gorm:"column:C_ENDTIME"`
	Announcement   string    `gorm:"column:C_ANNOUNCEMENT"`
	Password       *string   `gorm:"column:C_PASSWORD"`
	ReplayStatusId int       `gorm:"column:C_REPLAY_STATUS_ID"`
	ManagerId      int       `gorm:"column:C_MANAGER_ID"`
}

func (VhojContest) TableName() string {
	return "t_contest"
}

type VhojContestProblem struct {
	ContestID int    `gorm:"column:C_CONTEST_ID"`
	ProblemID string `gorm:"column:C_PROBLEM_ID"`
	Num       string `gorm:"column:C_NUM"`
}

func (VhojContestProblem) TableName() string {
	return "t_cproblem"
}

func (s *MigrateContestService) Start() error {
	ctx := context.Background()

	slog.Info("migrate contest start")

	var contests []*foundationmodel.Contest

	jolContests, err := s.processJolContest(ctx)
	if err != nil {
		return metaerror.Wrap(err, "process JolContest failed")
	}
	contests = append(contests, jolContests...)

	vhojContests, err := s.processVhojContest(ctx)
	if err != nil {
		return metaerror.Wrap(err, "process VhojContest failed")
	}
	contests = append(contests, vhojContests...)

	slog.Info("migrate contest", slog.Int("length", len(contests)))

	sort.Slice(
		contests, func(i, j int) bool {
			return contests[i].CreateTime.Before(contests[j].CreateTime)
		},
	)

	for i, contest := range contests {
		if contest.MigrateJolId > 0 {
			s.oldJolContestIdToNewContestId[contest.MigrateJolId] = i + 1
		}
		if contest.MigrateVhojId > 0 {
			s.oldVhojContestIdToNewContestId[contest.MigrateVhojId] = i + 1
		}
	}

	for _, contest := range contests {
		err := foundationdao.GetContestDao().InsertContest(ctx, contest)
		if err != nil {
			return metaerror.Wrap(err, "insert contest failed")
		}
	}

	slog.Info("migrate contest success")

	return nil
}

func (s *MigrateContestService) processJolContest(ctx context.Context) ([]*foundationmodel.Contest, error) {
	slog.Info("migrate JolContest processJolContest")

	db := metamysql.GetSubsystem().GetClient("jol")

	var contests []JolContest
	err := db.Raw(`select *  from contest left join (select * from privilege where rightstr like 'm%') p on concat('m',contest_id)=rightstr order by contest_id asc`).
		Scan(&contests).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "query contests failed")
	}

	var contestDocs []*foundationmodel.Contest

	for _, p := range contests {

		userId, err := GetMigrateUserService().getUserIdByUsername(ctx, p.UserId)
		if err != nil {
			return nil, err
		}

		var problems []JolContestProblem
		if err := db.WithContext(ctx).
			Where("contest_id = ?", p.ContestID).
			Find(&problems).Error; err != nil {
			return nil, err
		}

		finalContest := foundationmodel.NewContestBuilder().
			Title(p.Title).
			Description(p.Description).
			Notification(p.Notification).
			StartTime(p.StartTime).
			EndTime(p.EndTime).
			Password(p.Password).
			Type(foundationenum.ContestType(p.Kind)).
			OwnerId(userId).
			Build()

		for _, problem := range problems {
			newProblemId := GetMigrateProblemService().GetNewProblemId(problem.ProblemID)
			finalContest.Problems = append(
				finalContest.Problems,
				foundationmodel.NewContestProblemBuilder().
					ProblemId(newProblemId).
					Weight(problem.Scores).
					Index(problem.Num+1).
					Build(),
			)
		}

		finalContest.CreateTime = p.StartTime

		finalContest.MigrateJolId = p.ContestID

		contestDocs = append(contestDocs, finalContest)
	}

	return contestDocs, nil
}

func (s *MigrateContestService) processVhojContest(ctx context.Context) ([]*foundationmodel.Contest, error) {
	slog.Info("migrate JolContest processVhojContest")

	db := metamysql.GetSubsystem().GetClient("vhoj")
	var contests []VhojContest
	if err := db.Order("C_ID ASC").Find(&contests).Error; err != nil {
		return nil, metaerror.Wrap(err, "query contests failed")
	}
	var contestDocs []*foundationmodel.Contest
	for _, p := range contests {
		owner, err := GetMigrateUserService().getUsernameByVhojId(p.ManagerId)
		if err != nil {
			return nil, metaerror.Wrap(err, "get owner failed")
		}
		ownerUserId, err := GetMigrateUserService().getUserIdByUsername(ctx, owner)
		if err != nil {
			return nil, metaerror.Wrap(err, "get owner user id failed")
		}
		var problems []VhojContestProblem
		if err := db.WithContext(ctx).
			Where("C_CONTEST_ID = ?", p.ContestID).
			Find(&problems).Error; err != nil {
			return nil, err
		}

		finalContest := foundationmodel.NewContestBuilder().
			Title(p.Title).
			Description(p.Description).
			Notification(p.Announcement).
			StartTime(p.BeginTime).
			EndTime(p.EndTime).
			Password(p.Password).
			OwnerId(ownerUserId).
			Build()

		sort.Slice(
			problems, func(i, j int) bool {
				return problems[i].Num < problems[j].Num
			},
		)

		for i, problem := range problems {

			var vhojProblem VhojProblem
			err := db.Where("C_ID = ?", problem.ProblemID).First(&vhojProblem).Error
			if err != nil {
				return nil, metaerror.Wrap(err, "query problems failed")
			}

			var newProblemId string

			if vhojProblem.OriginOj == "HPU" {
				hpuId, err := strconv.Atoi(vhojProblem.OriginProb)
				if err != nil {
					return nil, metaerror.Wrap(err, "parse hpu id failed")
				}
				newProblemId = GetMigrateProblemService().GetNewProblemId(hpuId)
			} else {
				newProblemId = fmt.Sprintf("%s-%s", vhojProblem.OriginOj, vhojProblem.OriginProb)
			}

			finalContest.Problems = append(
				finalContest.Problems,
				foundationmodel.NewContestProblemBuilder().
					ProblemId(newProblemId).
					Index(i+1).
					Build(),
			)
		}

		finalContest.CreateTime = p.BeginTime

		finalContest.MigrateVhojId = p.ContestID

		contestDocs = append(contestDocs, finalContest)
	}
	return contestDocs, nil
}

func (s *MigrateContestService) GetNewContestIdByJol(oldContestId int) int {
	if newContestId, ok := s.oldJolContestIdToNewContestId[oldContestId]; ok {
		return newContestId
	}
	return 0
}

func (s *MigrateContestService) GetNewContestIdByVhoj(oldContestId int) int {
	if newContestId, ok := s.oldVhojContestIdToNewContestId[oldContestId]; ok {
		return newContestId
	}
	return 0
}
