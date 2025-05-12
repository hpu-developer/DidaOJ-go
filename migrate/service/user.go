package service

import (
	"context"
	"log/slog"
	"time"

	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type MigrateUserService struct{}

var singletonMigrateUserService = singleton.Singleton[MigrateUserService]{}

func GetMigrateUserService() *MigrateUserService {
	return singletonMigrateUserService.GetInstance(
		func() *MigrateUserService {
			return &MigrateUserService{}
		},
	)
}

// GORM 模型定义
type JolUser struct {
	UserID   string    `gorm:"column:user_id"`
	Nick     string    `gorm:"column:nick"`
	Password string    `gorm:"column:password"`
	Email    string    `gorm:"column:email"`
	Exper    int       `gorm:"column:exper"`
	Sign     string    `gorm:"column:sign"`
	School   string    `gorm:"column:school"`
	RegTime  time.Time `gorm:"column:reg_time"`
	VjudgeId string    `gorm:"column:vjudge_id"`
}

func (JolUser) TableName() string {
	return "Users"
}

type CodeojUser struct {
	UserID       string    `gorm:"column:user_id"`
	Nickname     string    `gorm:"column:nickname"`
	Password     string    `gorm:"column:password"`
	Email        string    `gorm:"column:email"`
	Sign         string    `gorm:"column:sign"`
	Organization string    `gorm:"column:organization"`
	RegTime      time.Time `gorm:"column:reg_time"`
}

func (CodeojUser) TableName() string {
	return "User"
}

func (s *MigrateUserService) Start() error {
	ctx := context.Background()

	if err := s.processJolUser(ctx); err != nil {
		return err
	}
	if err := s.processCodeojUser(ctx); err != nil {
		return err
	}

	return nil
}

func (s *MigrateUserService) processJolUser(ctx context.Context) error {
	slog.Info("migrate User processJolUser")

	db := metamysql.GetSubsystem().GetClient("jol")

	var users []JolUser
	if err := db.Order("reg_time ASC").Find(&users).Error; err != nil {
		return metaerror.Wrap(err, "query Users failed")
	}

	var docs []*foundationmodel.User
	for _, u := range users {
		seq, err := foundationdao.GetCounterDao().GetNextSequence(ctx, "user_id")
		if err != nil {
			return err
		}

		user := foundationmodel.NewUserBuilder().
			Id(seq).
			Username(u.UserID).
			Nickname(u.Nick).
			Password(u.Password).
			Email(u.Email).
			CheckinCount(u.Exper).
			Sign(u.Sign).
			Organization(u.School).
			VjudgeId(u.VjudgeId).
			RegTime(u.RegTime).
			Accept(0).
			Attempt(0).
			Build()

		docs = append(docs, user)
	}

	if len(docs) > 0 {
		if err := foundationdao.GetUserDao().UpdateUsers(ctx, docs); err != nil {
			return err
		}
		slog.Info("migrate User processJolUser success")
	}

	return nil
}

func (s *MigrateUserService) processCodeojUser(ctx context.Context) error {
	slog.Info("migrate User processCodeojUser")

	db := metamysql.GetSubsystem().GetClient("codeoj")

	var users []CodeojUser
	if err := db.Order("reg_time ASC").Find(&users).Error; err != nil {
		return metaerror.Wrap(err, "query User failed")
	}

	var docs []*foundationmodel.User
	for _, u := range users {
		user := foundationmodel.NewUserBuilder().
			Username(u.UserID).
			Nickname(u.Nickname).
			Password(u.Password).
			Email(u.Email).
			Sign(u.Sign).
			Organization(u.Organization).
			RegTime(u.RegTime).
			Build()

		if user.Username == "BoilTask" {
			user.Roles = []string{"r-admin"}
		}

		docs = append(docs, user)
	}

	if len(docs) > 0 {
		if err := foundationdao.GetUserDao().UpdateUsers(ctx, docs); err != nil {
			return err
		}
		slog.Info("migrate User processCodeojUser success")
	}

	return nil
}
