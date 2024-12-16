package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"wgxDouYin/cmd/api/handler"
	"wgxDouYin/cmd/api/rpc"
	"wgxDouYin/pkg/middleware"
	"wgxDouYin/pkg/viper"
)

var (
	apiConfig     = viper.Init("api")
	apiServerName = apiConfig.Viper.GetString("server.name")
	apiServerAddr = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("server.host"), apiConfig.Viper.GetInt("server.port"))
	skipRoutes    = []string{
		"/wgxDouYin/user/register/",
		"/wgxDouYin/user/login/",
	}
)

func InitRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/wgxDouYin")
	v1.Use(middleware.TokenAuthMiddleware(rpc.KeysManager, skipRoutes...))
	{
		user := v1.Group("/user")
		{
			user.POST("/register/", handler.UserRegister)
			user.POST("/login/", handler.UserLogin)
			user.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, map[string]string{
					"msg": "test requests.",
				})
			})
		}
	}
	return router
}

func main() {
	r := InitRouter()
	server := &http.Server{
		Addr:           apiServerAddr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("gateway启动失败,err:%v\n", err)
	}
}
