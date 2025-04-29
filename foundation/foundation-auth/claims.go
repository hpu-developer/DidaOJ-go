package foundationauth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	jwt.RegisteredClaims
	UserId int `json:"user_id"`
}

func (c *Claims) IsValid() bool {
	return c != nil && c.UserId > 0
}
