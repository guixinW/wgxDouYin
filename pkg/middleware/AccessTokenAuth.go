package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"wgxDouYin/internal/response"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/jwt"
	"wgxDouYin/pkg/keys"
)

// AccessTokenAuthMiddleware JWT验证中间件.skipRoutes为无需验证的请求
func AccessTokenAuthMiddleware(serviceDependencyMap map[string]string, keys *keys.KeyManager, skipRoutes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//queryUserId, err := strconv.ParseUint(c.Query("query_user_id"), 10, 64)
		//if queryUserId == 3 {
		//	c.Set("token_user_id", uint64(2))
		//	c.Next()
		//}
		for _, skipRoute := range skipRoutes {
			if 0 == strings.Compare(c.FullPath(), skipRoute) {
				c.Next()
				return
			}
		}
		serviceName, err := getServiceName(c.FullPath())
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, err.Error())
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusTokenInvalid, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
			return
		}
		authParts := strings.Split(authHeader, " ")
		if authParts[0] != "Bearer" || len(authParts) != 2 {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, "Authorization header format is incorrect")
			return
		}
		AccessToken := authParts[1]
		dependencyService := serviceDependencyMap[serviceName]
		publicKey, err := keys.GetServerPublicKey(etcd.KeyPrefix(dependencyService))
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, err.Error())
			return
		}
		if publicKey == nil {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, "public key is nil")
			return
		}
		claim, err := jwt.ParseToken(publicKey, AccessToken)
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusTokenExpired, err.Error())
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
