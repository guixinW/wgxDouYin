package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"wgxDouYin/internal/response"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/jwt"
	"wgxDouYin/pkg/keys"
)

func RefreshTokenMiddleware(serviceDependencyMap map[string]string, keys *keys.KeyManager, RefreshTokenRouter string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if 0 != strings.Compare(c.FullPath(), RefreshTokenRouter) {
			c.Next()
			return
		}
		serviceName, err := getServiceName(c.FullPath())
		if err != nil {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, err.Error())
		}
		RefreshToken, err := c.Cookie("refresh_token")
		fmt.Printf("refresh_token:%v\n", RefreshToken)
		if err != nil || RefreshToken == "" {
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, "refresh token is nil")
			return
		}
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
		claim, err := jwt.ParseToken(publicKey, RefreshToken)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) && claim != nil {
				response.AbortWithError(c, http.StatusUnauthorized, response.StatusTokenExpired, "token expired")
				c.Abort()
				return
			}
			response.AbortWithError(c, http.StatusUnauthorized, response.StatusOther, err.Error())
			return
		}
		c.Set("token_user_id", claim.UserId)
		c.Set("refresh_token", RefreshToken)
		c.Next()
	}
}
