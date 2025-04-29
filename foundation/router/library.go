package router

import (
	foundationauth "foundation/foundation-auth"
	foundationconfig "foundation/foundation-config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"meta/auth"
	"meta/response"
	"net/http"
)

func AuthMiddlewareOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("X-Token")
		if token != "" {
			slog.Info("token", "token", token)
			var jwtClaims foundationauth.Claims
			err := auth.ValidateJWT(
				token, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
					return foundationconfig.GetJwtSecret(), nil
				},
			)
			if err == nil {
				if jwtClaims.IsValid() {
					c.Set("claims", jwtClaims)
				}
			}
		}
		c.Next()
	}
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("X-Token")
		if token == "" {
			response.NewResponse(c, http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		slog.Info("token", "token", token)

		var jwtClaims foundationauth.Claims
		err := auth.ValidateJWT(
			token, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
				return foundationconfig.GetJwtSecret(), nil
			},
		)
		if err != nil {
			response.NewResponse(c, http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		if !jwtClaims.IsValid() {
			response.NewResponse(c, http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		c.Set("claims", jwtClaims)
		c.Next()
	}
}
