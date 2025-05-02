package foundationservice

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	foundationauth "foundation/foundation-auth"
	foundationconfig "foundation/foundation-config"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	metaerror "meta/meta-error"
	"meta/singleton"
	"web/response"
)

type UserService struct {
}

var singletonUserService = singleton.Singleton[UserService]{}

func GetUserService() *UserService {
	return singletonUserService.GetInstance(
		func() *UserService {
			return &UserService{}
		},
	)
}

func (s *UserService) GetUser(ctx context.Context, userId int) (*foundationmodel.User, error) {
	return foundationdao.GetUserDao().GetUser(ctx, userId)
}

func (s *UserService) GetUserLoginResponse(ctx context.Context, userId int) (*response.UserLogin, error) {
	resultUser, err := foundationdao.GetUserDao().GetUserLogin(ctx, userId)
	if err != nil {
		return nil, err
	}
	if resultUser == nil {
		return nil, nil
	}
	token, err := s.GetTokenByUserId(resultUser.Id, foundationconfig.GetJwtSecret())
	if err != nil {
		return nil, err
	}
	userResponse := response.NewUserLoginBuilder().
		Token(*token).
		UserId(resultUser.Id).
		Username(resultUser.Username).
		Nickname(resultUser.Nickname).
		Build()
	return userResponse, nil
}

func (s *UserService) Login(ctx *gin.Context, username string, password string) (*response.UserLogin, error) {
	resultUser, err := foundationdao.GetUserDao().GetUserLoginByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if resultUser == nil {
		return nil, nil
	}
	hash := md5.New()
	_, err = hash.Write([]byte(password))
	if err != nil {
		return nil, metaerror.Wrap(err)
	}
	passwordInputMd5 := hex.EncodeToString(hash.Sum(nil))
	if passwordInputMd5 != resultUser.Password {
		return nil, nil
	}

	token, err := s.GetTokenByUserId(resultUser.Id, foundationconfig.GetJwtSecret())
	if err != nil {
		return nil, err
	}

	userResponse := response.NewUserLoginBuilder().
		Token(*token).
		UserId(resultUser.Id).
		Username(resultUser.Username).
		Nickname(resultUser.Nickname).
		Build()

	return userResponse, nil
}

func (s *UserService) GetTokenByUserId(userId int, secret []byte) (*string, error) {
	token, err := foundationauth.GetToken(userId, secret)
	if err != nil {
		return nil, err
	}
	return token, nil
}
