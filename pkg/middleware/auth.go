package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"wgxDouYin/pkg/jwt"
)

func getServerName(path string) string {
	if strings.HasPrefix(path, "/wgxDouYin/user/") {
		return "user"
	}
	return ""
}

// TokenAuthMiddleware JWT验证中间件.skipRoutes为无需验证的请求
func TokenAuthMiddleware(keys *jwt.KeyManager, skipRoutes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, skipRoute := range skipRoutes {
			if 0 == strings.Compare(c.FullPath(), skipRoute) {
				c.Next()
				return
			}
		}
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
		fmt.Printf("server path:%v\n", c.Request.URL.Path)
		serverName := getServerName(c.Request.URL.Path)
		fmt.Printf("server name:%v\n", serverName)
		claim, err := keys.ParseToken(tokenString, serverName)
		if err != nil {
			responseWithError(c, http.StatusUnauthorized, err)
			return
		}
		c.Set("UserID", claim.UserId)
		c.Next()
	}
}
