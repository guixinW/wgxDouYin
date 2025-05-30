package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "net/http/pprof" // 注册 pprof 的 handler
	"time"
	"wgxDouYin/cmd/api/handler"
	"wgxDouYin/cmd/api/rpc"
	"wgxDouYin/pkg/middleware"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		path := c.FullPath() // 路由路径，不是实际 URL，避免高维度
		if path == "" {
			path = "unknown"
		}
		requestCounter.WithLabelValues(c.Request.Method, path, string(rune(status))).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

var (
	apiConfig         = viper.Init("api")
	apiServerAddr     = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("service.host"), apiConfig.Viper.GetInt("service.port"))
	accessSkipRouters = []string{
		"/wgxdouyin/user/register/",
		"/wgxdouyin/user/login/",
		"/wgxdouyin/user/refreshToken/",
		"/wgxdouyin/user/loginPage/",
	}
	refreshRouter        = "/wgxdouyin/user/refreshToken/"
	ServiceNameMap       map[string]string
	ServiceDependencyMap map[string]string
	logger               = zap.InitLogger()
)

func init() {
	if err := apiConfig.Viper.UnmarshalKey("otherService", &ServiceNameMap); err != nil {
		panic(err)
	}
	if err := apiConfig.Viper.UnmarshalKey("serviceDependency", &ServiceDependencyMap); err != nil {
		panic(err)
	}
	prometheus.MustRegister(requestCounter, requestDuration)
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	//利用nginx替代cors
	//router.Use(cors.New(cors.Config{
	//	AllowOrigins:     []string{"http://localhost:8080"},
	//	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	//	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	//	AllowCredentials: true,
	//}))

	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}
	wgxDouYin := router.Group("/wgxdouyin")
	wgxDouYin.Use(
		PrometheusMiddleware(),
		middleware.ServiceAvailabilityMiddleware(ServiceNameMap, rpc.KeysManager),
		middleware.AccessTokenAuthMiddleware(ServiceDependencyMap, rpc.KeysManager, accessSkipRouters...),
		middleware.RefreshTokenMiddleware(ServiceNameMap, rpc.KeysManager, refreshRouter))
	{
		user := wgxDouYin.Group("/user")
		{
			user.GET("/refreshToken/", handler.RefreshToken)
			user.POST("/register/", handler.UserRegister)
			user.POST("/login/", handler.UserLogin)
			user.GET("/userInform/", handler.UserInform)
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
	r := InitRouter()
	if err := r.Run(apiServerAddr); err != nil {
		panic(err)
	}
}
