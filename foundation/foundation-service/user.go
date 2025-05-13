package foundationservice

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
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
		Roles(resultUser.Roles).
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
		decoded, err := base64.StdEncoding.DecodeString(resultUser.Password)
		if err != nil {
			return nil, metaerror.Wrap(err)
		}
		if len(decoded) <= 20 {
			return nil, metaerror.New("password decoded error", "len", len(decoded))
		}
		salt := decoded[20:]
		md5Hex := md5.Sum([]byte(password))
		md5HexStr := make([]byte, 32)
		hex.Encode(md5HexStr, md5Hex[:])
		sha1Hasher := sha1.New()
		sha1Hasher.Write(md5HexStr)
		sha1Hasher.Write(salt)
		sha1Hash := sha1Hasher.Sum(nil)
		final := append(sha1Hash, salt...)
		encoded := base64.StdEncoding.EncodeToString(final)
		if encoded != resultUser.Password {
			return nil, nil
		}
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
		Roles(resultUser.Roles).
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

func (s *UserService) CheckUserAuth(ctx *gin.Context, auth foundationauth.AuthType) (int, bool, error) {
	userId, err := foundationauth.GetUserIdFromContext(ctx)
	if err != nil {
		return 0, false, nil
	}
	ok, err := s.CheckUserAuthByUserId(ctx, userId, foundationauth.AuthTypeManageProblem)
	if err != nil {
		return 0, false, err
	}
	if !ok {
		return 0, false, nil
	}
	return userId, true, nil
}

func (s *UserService) CheckUserAuthByUserId(ctx context.Context, userId int, auth foundationauth.AuthType) (bool, error) {
	userRoles, err := foundationdao.GetUserDao().GetUserRoles(ctx, userId)
	if err != nil {
		return false, err
	}
	ok := foundationconfig.CheckRolesHasAuth(userRoles, auth)
	return ok, nil
}

func (s *UserService) CheckUserAuthsByUserId(ctx context.Context, userId int, auths []foundationauth.AuthType) (bool, error) {
	userRoles, err := foundationdao.GetUserDao().GetUserRoles(ctx, userId)
	if err != nil {
		return false, err
	}
	ok := foundationconfig.CheckRolesHasAllAuths(userRoles, auths)
	return ok, nil
}
