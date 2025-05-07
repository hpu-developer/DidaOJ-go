package service

import (
	"context"
	"database/sql"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metapanic "meta/meta-panic"
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

func (s *MigrateUserService) Start() error {

	ctx := context.Background()

	err := s.processJolUser(ctx)
	if err != nil {
		return err
	}
	err = s.processCodeojUser(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *MigrateUserService) processJolUser(ctx context.Context) error {

	slog.Info("migrate User processJolUser")

	jolMysqlClient := metamysql.GetSubsystem().GetClient("jol")

	// User 定义
	type User struct {
		UserID   string
		Nickname sql.NullString
		Password sql.NullString
		Email    sql.NullString
		Exper    sql.NullInt32 // 签到次数
		Sign     sql.NullString
		School   sql.NullString
		RegTime  sql.NullTime
		VjudgeId sql.NullString
	}

	// === 拉取 User 表，按照reg_time排序，注册早的靠前 ===
	UserRows, err := jolMysqlClient.Query(`
		SELECT user_id, nick, password, email, exper, sign, school, reg_time, vjudge_id 
		FROM Users
		ORDER BY reg_time ASC
	`)
	if err != nil {
		return metaerror.Wrap(err, "query User row failed")
	}
	defer func(UserRows *sql.Rows) {
		err := UserRows.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(UserRows)

	// === 处理每一条 User 并插入 MongoDB ===
	var UserDocs []*foundationmodel.User

	for UserRows.Next() {
		var p User
		if err := UserRows.Scan(
			&p.UserID, &p.Nickname, &p.Password,
			&p.Email, &p.Exper, &p.Sign,
			&p.School, &p.RegTime, &p.VjudgeId,
		); err != nil {
			return metaerror.Wrap(err, "query User row failed")
		}

		seq, err := foundationdao.GetCounterDao().GetNextSequence(ctx, "user_id")
		if err != nil {
			return err
		}

		UserDocs = append(UserDocs, foundationmodel.NewUserBuilder().
			Id(seq).
			Username(p.UserID).
			Nickname(p.Nickname.String).
			Password(p.Password.String).
			Email(p.Email.String).
			Accept(0).
			Attempt(0).
			CheckinCount(int(p.Exper.Int32)).
			Sign(p.Sign.String).
			Organization(p.School.String).
			VjudgeId(p.VjudgeId.String).
			RegTime(metamysql.NullTimeToTime(p.RegTime)).
			Build())
	}

	// 插入 MongoDB
	if len(UserDocs) > 0 {
		//err = UserCol.Drop(ctx) // 清空原 User 集合
		//if err != nil {
		//	log.Fatal("清空 User 出错:", err)
		//}
		err = foundationdao.GetUserDao().UpdateUsers(ctx, UserDocs)
		if err != nil {
			return err
		}
	}

	slog.Info("migrate User processJolUser success")

	return nil
}

func (s *MigrateUserService) processCodeojUser(ctx context.Context) error {

	slog.Info("migrate User processCodeojUser")

	codeojMysqlClient := metamysql.GetSubsystem().GetClient("codeoj")

	// User 定义
	type User struct {
		UserID       string
		Nickname     sql.NullString
		Password     sql.NullString
		Email        sql.NullString
		Sign         sql.NullString
		Organization sql.NullString
		RegTime      sql.NullTime
	}

	// === 拉取 User 表，按照reg_time排序，注册早的靠前 ===
	UserRows, err := codeojMysqlClient.Query(`
		SELECT user_id, nickname, password, email, sign, organization, reg_time 
		FROM User
		ORDER BY reg_time ASC
	`)
	if err != nil {
		return metaerror.Wrap(err, "query User row failed")
	}
	defer func(UserRows *sql.Rows) {
		err := UserRows.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(UserRows)

	// === 处理每一条 User 并插入 MongoDB ===
	var UserDocs []*foundationmodel.User

	for UserRows.Next() {
		var p User
		if err := UserRows.Scan(
			&p.UserID, &p.Nickname, &p.Password,
			&p.Email, &p.Sign, &p.Organization,
			&p.RegTime,
		); err != nil {
			return metaerror.Wrap(err, "query User row failed")
		}

		newUser := foundationmodel.NewUserBuilder().
			Username(p.UserID).
			Nickname(p.Nickname.String).
			Password(p.Password.String).
			Email(p.Email.String).
			Sign(p.Sign.String).
			Organization(p.Organization.String).
			RegTime(metamysql.NullTimeToTime(p.RegTime)).
			Build()

		if newUser.Username == "BoilTask" {
			newUser.Roles = []string{"r-admin"}
		}

		userId, err := foundationdao.GetUserDao().GetUserIdByUsername(ctx, p.UserID)
		if err != nil {
			seq, err := foundationdao.GetCounterDao().GetNextSequence(ctx, "user_id")
			if err != nil {
				return err
			}
			newUser.Id = seq
			UserDocs = append(UserDocs, newUser)
		} else {
			newUser.Id = userId
			err = foundationdao.GetUserDao().UpdateUser(ctx, userId, newUser)
			if err != nil {
				return err
			}
		}
	}

	// 插入 MongoDB
	if len(UserDocs) > 0 {
		//err = UserCol.Drop(ctx) // 清空原 User 集合
		//if err != nil {
		//	log.Fatal("清空 User 出错:", err)
		//}

		err = foundationdao.GetUserDao().UpdateUsers(ctx, UserDocs)
		if err != nil {
			return err
		}
	}

	slog.Info("migrate User processCodeojUser success")

	return nil
}
