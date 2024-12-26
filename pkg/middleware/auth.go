package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/jwt"
	"wgxDouYin/pkg/keys"
)

// TokenAuthMiddleware JWT验证中间件.skipRoutes为无需验证的请求
func TokenAuthMiddleware(serviceNameMap map[string]string, keys *keys.KeyManager, skipRoutes ...string) gin.HandlerFunc {
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
		publicKey, err := keys.GetServerPublicKey(etcd.KeyPrefix(serviceNameMap[c.Request.URL.Path]))
		if err != nil {
			responseWithError(c, http.StatusUnauthorized, err.Error())
			return
		}
		if publicKey == nil {
			responseWithError(c, http.StatusUnauthorized, "public key is nil")
			return
		}
		claim, err := jwt.ParseToken(publicKey, tokenString)
		if err != nil {
			responseWithError(c, http.StatusUnauthorized, err.Error())
			return
		}
		c.Set("UserID", claim.UserId)
		c.Next()
	}
}
