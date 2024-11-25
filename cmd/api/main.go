package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"wgxDouYin/cmd/api/handler"
	"wgxDouYin/pkg/viper"
)

var (
	apiConfig     = viper.Init("api")
	apiServerName = apiConfig.Viper.GetString("server.name")
	apiServerAddr = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("server.host"), apiConfig.Viper.GetInt("server.port"))
)

func InitRouter() *gin.Engine {
	//logger := zap.InitLogger()
	router := gin.Default()
	v1 := router.Group("/wgxDouYin")
	{
		user := v1.Group("/user")
		{
			user.POST("/register/", handler.UserRegister)
			user.POST("/login/", handler.UserLogin)
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
