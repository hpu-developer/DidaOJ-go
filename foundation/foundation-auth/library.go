package foundationauth

import (
	"meta/auth"
	metaerror "meta/meta-error"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetToken(userId int, nowTime time.Time, secret []byte) (*string, error) {
	duration := time.Hour * 24 * 7
	return GetTokenExpiration(userId, nowTime, duration, secret)
}

func GetTokenExpiration(userId int, nowTime time.Time, duration time.Duration, secret []byte) (
	*string,
	error,
) {
	expirationTime := nowTime.Add(duration)
	var claims = &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		UserId: userId,
	}
	return auth.GetToken(claims, secret)
}

func GetUserIdFromContext(c *gin.Context) (int, error) {
	claimsPtr := c.Value("claims")
	if claimsPtr == nil {
		return -1, metaerror.New("claims not found in context")
	}
	claims, err := claimsPtr.(Claims)
	if !err {
		return -1, metaerror.New("claims type error")
	}
	return claims.UserId, nil
}
