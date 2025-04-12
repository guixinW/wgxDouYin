package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/keys"
)

func ServiceAvailabilityMiddleware(ServiceNameMap map[string]string, keys *keys.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName, err := getServiceName(c.FullPath())
		keys.PrintServiceAndKey()
		_, err = keys.GetServerPublicKey(etcd.KeyPrefix(ServiceNameMap[serviceName]))
		if err != nil {
			fmt.Println(err)
			responseWithError(c, http.StatusUnauthorized, "功能未上线")
			return
		}
		c.Next()
	}
}
