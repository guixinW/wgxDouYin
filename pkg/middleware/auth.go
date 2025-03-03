package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/jwt"
	"wgxDouYin/pkg/keys"
)

// TokenAuthMiddleware JWT验证中间件.skipRoutes为无需验证的请求
func TokenAuthMiddleware(serviceDependencyMap map[string]string, keys *keys.KeyManager, skipRoutes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName, err := getServiceName(c.FullPath())
		fmt.Println(serviceName)
		if err != nil {
			responseWithError(c, http.StatusUnauthorized, err)
		}
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
		dependencyService := serviceDependencyMap[serviceName]
		publicKey, err := keys.GetServerPublicKey(etcd.KeyPrefix(dependencyService))
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
		c.Set("token_user_id", claim.UserId)
		c.Next()
	}
}

func getServiceName(path string) (string, error) {
	trimmedPath := strings.Trim(path, "/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid path format: %s", path)
	}
	serviceName := parts[1]
	return serviceName, nil
}
