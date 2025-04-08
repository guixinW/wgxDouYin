package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/keys"
)

func ServiceAvailabilityMiddleware(ServiceNameMap map[string]string, keys *keys.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName, err := getServiceName(c.FullPath())
		_, err = keys.GetServerPublicKey(etcd.KeyPrefix(ServiceNameMap[serviceName]))
		if err != nil {
			responseWithError(c, http.StatusUnauthorized, "功能未上线")
			return
		}
		c.Next()
	}
}
