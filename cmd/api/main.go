package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "net/http/pprof" // 注册 pprof 的 handler
	"wgxDouYin/cmd/api/handler"
	"wgxDouYin/cmd/api/rpc"
	"wgxDouYin/pkg/middleware"
	"wgxDouYin/pkg/viper"
)

var (
	apiConfig     = viper.Init("api")
	apiServerAddr = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("service.host"), apiConfig.Viper.GetInt("service.port"))
	skipRoutes    = []string{
		"/wgxdouyin/user/register/",
		"/wgxdouyin/user/login/",
	}
	ServiceNameMap       map[string]string
	ServiceDependencyMap map[string]string
	//logger               = zap.InitLogger()
)

func init() {
	if err := apiConfig.Viper.UnmarshalKey("otherService", &ServiceNameMap); err != nil {
		panic(err)
	}
	if err := apiConfig.Viper.UnmarshalKey("serviceDependency", &ServiceDependencyMap); err != nil {
		panic(err)
	}
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}
	wgxDouYin := router.Group("/wgxdouyin")
	wgxDouYin.Use(
		middleware.ServiceAvailabilityMiddleware(ServiceNameMap, rpc.KeysManager),
		middleware.TokenAuthMiddleware(ServiceDependencyMap, rpc.KeysManager, skipRoutes...))
	{
		user := wgxDouYin.Group("/user")
		{
			user.POST("/register/", handler.UserRegister)
			user.POST("/login/", handler.UserLogin)
			user.GET("/", handler.UserInform)
		}
		relation := wgxDouYin.Group("/relation")
		{
			relation.POST("/action/", handler.RelationAction)
			relation.POST("/friend/list/", handler.FriendList)
			relation.POST("/following/list/", handler.FollowingList)
			relation.POST("/follower/list/", handler.FollowerList)
		}
		publish := wgxDouYin.Group("/video")
		{
			publish.GET("/feed/", handler.Feed)
			publish.GET("/list/", handler.PublishList)
			publish.POST("/action/", handler.PublishAction)
			publish.GET("/generatePostURL/", handler.PublishPostURL)
		}
		favorite := wgxDouYin.Group("/favorite")
		{
			favorite.POST("/action/", handler.FavoriteAction)
			favorite.POST("/list/", handler.FavoriteList)
		}
		comment := wgxDouYin.Group("/comment")
		{
			comment.POST("/action/", handler.CommentAction)
			comment.GET("/list/", handler.CommentList)
		}
	}
	return router
}

func main() {
	//go func() {
	//	fmt.Println("pprof 监听在 http://localhost:6060/debug/pprof/")
	//	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
	//		panic("pprof 启动失败: " + err.Error())
	//	}
	//}()
	r := InitRouter()
	if err := r.Run(apiServerAddr); err != nil {
		panic(err)
	}
}
