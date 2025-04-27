package foundationauth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"meta/auth"
	metaerror "meta/meta-error"
	goTime "time"
)

func GetToken(userId string, openId string, secret []byte) (*string, error) {
	duration := goTime.Hour * 24 * 7
	return GetTokenExpiration(userId, openId, duration, secret)
}

func GetTokenExpiration(userId string, openId string, duration goTime.Duration, secret []byte) (
	*string,
	error,
) {
	expirationTime := goTime.Now().Add(duration)
	var claims = &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		UserId: userId,
	}
	return auth.GetToken(claims, secret)
}

func GetUserIdFromContext(c *gin.Context) (*string, error) {
	claimsPtr := c.Value("claims")
	if claimsPtr == nil {
		return nil, metaerror.New("claims not found in context")
	}
	claims, err := claimsPtr.(Claims)
	if !err {
		return nil, metaerror.New("claims type error")
	}
	return &claims.UserId, nil
}
