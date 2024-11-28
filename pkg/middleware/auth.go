package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"wgxDouYin/pkg/jwt"
)

func TokenAuthMiddleware(keys jwt.KeyManager, serverName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			responseWithError(c, http.StatusUnauthorized, "Authorization header is missing")
			return
		}

		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			responseWithError(c, http.StatusUnauthorized, "Authorization header format is incorrect")
			return
		}
		tokenString := authParts[1]
		token, err := keys.ParseToken(tokenString, serverName)
		if err != nil {
			responseWithError(c, http.StatusUnauthorized, err)
			return
		}
	}
}
