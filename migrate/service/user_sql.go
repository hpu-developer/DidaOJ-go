package service

import (
	"context"
	foundationdao "foundation/foundation-dao"
	foundationdaomongo "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	"meta/singleton"
)

type MigrateUserSqlService struct {
}

var singletonMigrateUserSqlService = singleton.Singleton[MigrateUserSqlService]{}

func GetMigrateUserSqlService() *MigrateUserSqlService {
	return singletonMigrateUserSqlService.GetInstance(
		func() *MigrateUserSqlService {
			return &MigrateUserSqlService{}
		},
	)
}

func (s *MigrateUserSqlService) Start(ctx context.Context) error {

	userList, err := foundationdaomongo.GetUserDao().GetUserListAll(ctx)
	if err != nil {
		return err
	}
	slog.Info("userList", "userList", len(userList))

	for _, user := range userList {
		newUser := foundationmodel.NewUserBuilder().
			Username(user.Username).
			Nickname(user.Nickname).
			RealName(user.RealName).
			Password(user.Password).
			Email(user.Email).
			Gender(user.Gender).
			Number(user.Number).
			Slogan(user.Slogan).
			Organization(user.Organization).
			QQ(user.QQ).
			VjudgeId(user.VjudgeId).
			Github(user.Github).
			Codeforces(user.Codeforces).
			CheckInCount(user.CheckinCount).
			InsertTime(user.RegTime).
			ModifyTime(user.RegTime).
			Accept(user.Accept).
			Attempt(user.Attempt).
			Build()
		err := foundationdao.GetUserDao().InsertUser(ctx, newUser)
		if err != nil {
			return err
		}
	}

	return nil
}
