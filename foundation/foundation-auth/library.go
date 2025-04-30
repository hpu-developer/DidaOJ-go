package foundationauth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"meta/auth"
	metaerror "meta/meta-error"
	goTime "time"
)

func GetToken(userId int, secret []byte) (*string, error) {
	duration := goTime.Hour * 24 * 7
	return GetTokenExpiration(userId, duration, secret)
}

func GetTokenExpiration(userId int, duration goTime.Duration, secret []byte) (
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
