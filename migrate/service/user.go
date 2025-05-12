package service

import (
	"context"
	"log/slog"
	"sort"
	"time"

	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type MigrateUserService struct {
	usernameToUserId map[string]int
}

var singletonMigrateUserService = singleton.Singleton[MigrateUserService]{}

func GetMigrateUserService() *MigrateUserService {
	return singletonMigrateUserService.GetInstance(
		func() *MigrateUserService {
			s := &MigrateUserService{}
			s.usernameToUserId = make(map[string]int)
			return s
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
	return "user"
}

func (s *MigrateUserService) Start() error {
	ctx := context.Background()

	var users []*foundationmodel.User

	jolUsers, err := s.processJolUser(ctx)
	if err != nil {
		return err
	}
	users = append(users, jolUsers...)

	usernameToUser := make(map[string]*foundationmodel.User)
	for _, user := range users {
		usernameToUser[user.Username] = user
	}

	codeojUsers, err := s.processCodeojUser(ctx)
	if err != nil {
		return err
	}
	for _, u := range codeojUsers {
		if _, ok := usernameToUser[u.Username]; ok {
			usernameToUser[u.Username] = u
			continue
		}
		users = append(users, u)
	}

	slog.Info("migrate users updates", "count", len(users))

	sort.Slice(users, func(i, j int) bool {
		return users[i].RegTime.Before(users[j].RegTime)
	})

	for _, user := range users {
		err = foundationdao.GetUserDao().InsertUser(ctx, user)
		if err != nil {
			return metaerror.Wrap(err, "insert user failed")
		}
	}

	return nil
}

func (s *MigrateUserService) processJolUser(ctx context.Context) ([]*foundationmodel.User, error) {
	slog.Info("migrate User processJolUser")

	db := metamysql.GetSubsystem().GetClient("jol")

	var users []JolUser
	if err := db.Order("reg_time ASC").Find(&users).Error; err != nil {
		return nil, metaerror.Wrap(err, "query Users failed")
	}

	var docs []*foundationmodel.User
	for _, u := range users {
		finalUser := foundationmodel.NewUserBuilder().
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

		docs = append(docs, finalUser)
	}

	return docs, nil
}

func (s *MigrateUserService) processCodeojUser(ctx context.Context) ([]*foundationmodel.User, error) {
	slog.Info("migrate User processCodeojUser")

	db := metamysql.GetSubsystem().GetClient("codeoj")

	var users []CodeojUser
	if err := db.Order("reg_time ASC").Find(&users).Error; err != nil {
		return nil, metaerror.Wrap(err, "query User failed")
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

	return docs, nil
}

func (s *MigrateUserService) getUserIdByUsername(ctx context.Context, username string) (int, error) {
	var err error
	userId, ok := s.usernameToUserId[username]
	if !ok {
		userId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, username)
		if err != nil {
			return -1, err
		}
		s.usernameToUserId[username] = userId
	}
	return userId, nil
}
