package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"wgxDouYin/cmd/api/handler"
	"wgxDouYin/cmd/api/rpc"
	"wgxDouYin/pkg/middleware"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	apiConfig     = viper.Init("api")
	apiServerAddr = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("service.host"), apiConfig.Viper.GetInt("service.port"))
	skipRoutes    = []string{
		"/wgxdouyin/user/register/",
		"/wgxdouyin/user/login/",
	}
	ServiceNameMap map[string]string
	logger         = zap.InitLogger()
)

func init() {
	if err := apiConfig.Viper.UnmarshalKey("otherService", &ServiceNameMap); err != nil {
		panic(err)
	}
	fmt.Printf("service name map:%v\n", ServiceNameMap)
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}
	v1 := router.Group("/wgxdouyin")
	v1.Use(middleware.TokenAuthMiddleware(ServiceNameMap, rpc.KeysManager, skipRoutes...))
	{
		user := v1.Group("/user")
		{
			user.POST("/register/", handler.UserRegister)
			user.POST("/login/", handler.UserLogin)
			user.GET("/", handler.UserInform)
		}
	}
	return router
}

func main() {
	r := InitRouter()
	if err := r.Run(apiServerAddr); err != nil {
		panic(err)
	}
}
