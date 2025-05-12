package service

import (
	"context"
	"log/slog"
	metamysql "meta/meta-mysql"
	"sort"
	"time"

	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
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

// GORM 模型定义
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

type Vhojontest struct {
	ContestID      int       `gorm:"column:C_ID"`
	Title          string    `gorm:"column:C_TITLE"`
	Description    string    `gorm:"column:C_DESCRIPTION"`
	BeginTime      time.Time `gorm:"column:C_BEGIN_TIME"`
	EndTime        time.Time `gorm:"column:C_END_TIME"`
	Announcement   string    `gorm:"column:C_ANNOUNCEMENT"`
	Password       *string   `gorm:"column:C_PASSWORD"`
	ReplayStatusId int       `gorm:"column:C_REPLAY_STATUS_ID"`
	ManagerId      int       `gorm:"column:C_MANAGER_ID"`
}

func (JolContest) TableName() string {
	return "t_contest"
}

func (s *MigrateContestService) Start() error {
	ctx := context.Background()

	slog.Info("migrate contest start")

	var contests []*foundationmodel.Contest
	contests, err := s.processJolContest(ctx)
	if err != nil {
		return metaerror.Wrap(err, "process JolContest failed")
	}

	slog.Info("migrate contest", slog.Int("length", len(contests)))

	sort.Slice(contests, func(i, j int) bool {
		return contests[i].CreateTime.Before(contests[j].CreateTime)
	})

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

		finalContest := foundationmodel.NewContestBuilder().
			Title(p.Title).
			Description(p.Description).
			Notification(p.Notification).
			StartTime(p.StartTime).
			EndTime(p.EndTime).
			Password(p.Password).
			Type(foundationmodel.ContestType(p.Kind)).
			OwnerId(userId).
			Build()

		finalContest.CreateTime = p.StartTime

		contestDocs = append(contestDocs, finalContest)
	}

	return contestDocs, nil
}

func (s *MigrateContestService) processVhojContest(ctx context.Context) ([]*foundationmodel.Contest, error) {
	slog.Info("migrate JolContest processVhojContest")

	db := metamysql.GetSubsystem().GetClient("vhoj")
	var contests []Vhojontest
	if err := db.Order("contest_id ASC").Find(&contests).Error; err != nil {
		return nil, metaerror.Wrap(err, "query contests failed")
	}
	var contestDocs []*foundationmodel.Contest
	for _, p := range contests {

		finalContest := foundationmodel.NewContestBuilder().
			Title(p.Title).
			Description(p.Description).
			Notification(p.Announcement).
			StartTime(p.BeginTime).
			EndTime(p.EndTime).
			Password(p.Password).
			Build()

		finalContest.CreateTime = p.BeginTime

		contestDocs = append(contestDocs, finalContest)
	}
	return contestDocs, nil
}
