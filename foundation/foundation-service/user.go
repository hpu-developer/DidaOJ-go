package foundationservice

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	foundationauth "foundation/foundation-auth"
	foundationconfig "foundation/foundation-config"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"foundation/foundation-request"
	foundationview "foundation/foundation-view"
	"io"
	metaerror "meta/meta-error"
	"meta/singleton"

	"github.com/gin-gonic/gin"
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

func (s *UserService) GetModifyInfo(ctx context.Context, userId int) (*foundationview.UserModifyInfo, error) {
	return foundationdao.GetUserDao().GetModifyInfo(ctx, userId)
}

func (s *UserService) GetInfoByUsername(ctx *gin.Context, username string) (*foundationview.UserInfo, error) {
	return foundationdao.GetUserDao().GetInfoByUsername(ctx, username)
}

func (s *UserService) GetUserLoginResponse(ctx context.Context, userId int) (*foundationview.UserLogin, error) {
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
	resultUser.Token = token
	resultUser.Roles, err = foundationdao.GetUserRoleDao().GetUserRoles(ctx, resultUser.Id)
	if err != nil {
		return nil, err
	}
	return resultUser, nil
}

func (s *UserService) GetEmailByUsername(ctx context.Context, username string) (*string, error) {
	return foundationdao.GetUserDao().GetEmailByUsername(ctx, username)
}

func (s *UserService) GetUserAccountInfo(ctx context.Context, userId int) (*foundationview.UserAccountInfo, error) {
	return foundationdao.GetUserDao().GetUserAccountInfo(ctx, userId)
}

func (s *UserService) GetUserAccountInfos(ctx context.Context, userIds []int) (
	[]*foundationview.UserAccountInfo, error,
) {
	return foundationdao.GetUserDao().GetUserAccountInfos(ctx, userIds)
}

func (s *UserService) GetUserAccountInfoByUsernames(
	ctx context.Context,
	usernames []string,
) ([]*foundationview.UserAccountInfo, error) {
	return foundationdao.GetUserDao().GetUserAccountInfosByUsername(ctx, usernames)
}

func (s *UserService) GetUserIdsByUsername(ctx *gin.Context, usernames []string) ([]int, error) {
	return foundationdao.GetUserDao().GetUserIdsByUsername(ctx, usernames)
}

func (s *UserService) InsertUser(ctx context.Context, user *foundationmodel.User) error {
	return foundationdao.GetUserDao().InsertUser(ctx, user)
}

func (s *UserService) GeneratePasswordEncode(password string) (string, error) {
	salt := make([]byte, 4)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	md5Hex := md5.Sum([]byte(password))
	md5HexStr := make([]byte, 32)
	hex.Encode(md5HexStr, md5Hex[:])
	sha1Hasher := sha1.New()
	sha1Hasher.Write(md5HexStr)
	sha1Hasher.Write(salt)
	sha1Hash := sha1Hasher.Sum(nil)
	final := append(sha1Hash, salt...)
	encoded := base64.StdEncoding.EncodeToString(final)
	return encoded, nil
}

func (s *UserService) UpdatePassword(ctx *gin.Context, username string, passwordEncode string) error {
	return foundationdao.GetUserDao().UpdatePassword(ctx, username, passwordEncode)
}

func (s *UserService) Login(ctx *gin.Context, username string, password string) (*foundationview.UserLogin, error) {
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
	resultUser.Roles, err = foundationdao.GetUserRoleDao().GetUserRoles(ctx, resultUser.Id)
	if err != nil {
		return nil, err
	}
	token, err := s.GetTokenByUserId(resultUser.Id, foundationconfig.GetJwtSecret())
	if err != nil {
		return nil, err
	}
	resultUser.Token = token
	return resultUser, nil
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
		return userId, false, nil
	}
	ok, err := s.CheckUserAuthByUserId(ctx, userId, auth)
	if err != nil {
		return userId, false, err
	}
	if !ok {
		return userId, false, nil
	}
	return userId, true, nil
}

func (s *UserService) CheckUserAuthByUserId(ctx context.Context, userId int, auth foundationauth.AuthType) (
	bool,
	error,
) {
	userRoles, err := foundationdao.GetUserRoleDao().GetUserRoles(ctx, userId)
	if err != nil {
		return false, err
	}
	ok := foundationconfig.CheckRolesHasAuth(userRoles, auth)
	return ok, nil
}

func (s *UserService) CheckUserAuthsByUserId(ctx context.Context, userId int, auths []foundationauth.AuthType) (
	bool,
	error,
) {
	userRoles, err := foundationdao.GetUserRoleDao().GetUserRoles(ctx, userId)
	if err != nil {
		return false, err
	}
	ok := foundationconfig.CheckRolesHasAllAuths(userRoles, auths)
	return ok, nil
}

func (s *UserService) GetRankAcAll(ctx *gin.Context, page int, pageSize int) ([]*foundationview.UserRank, int, error) {
	return foundationdao.GetUserDao().GetRankAcAll(ctx, page, pageSize)
}

func (s *UserService) FilterValidUserIds(ctx *gin.Context, userIds []int) ([]int, error) {
	return foundationdao.GetUserDao().FilterValidUserIds(ctx, userIds)
}

func (s *UserService) UpdateUserInfo(ctx context.Context, userId int, r *foundationrequest.UserModifyInfo) error {
	return foundationdao.GetUserDao().UpdateUserInfo(ctx, userId, r)
}
