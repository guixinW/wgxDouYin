package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"wgxDouYin/dal/db"
	userPb "wgxDouYin/grpc/user"
	app "wgxDouYin/internal/user/application/service"
	"wgxDouYin/internal/user/infrastructure/persistence"
	"wgxDouYin/internal/user/interfaces/grpc"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	config      = viper.Init("user")
	serviceName = config.Viper.GetString("service.name")
	serviceAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("service.host"),
		config.Viper.GetInt("service.port"))
	rpcAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("rpc.host"),
		config.Viper.GetInt("rpc.port"))
	etcdAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"),
		config.Viper.GetInt("etcd.port"))
	logger = zap.InitLogger()
)

func main() {
	// 初始化数据库
	db.Init()

	// 创建 DDD 组件
	userRepo := persistence.NewGormUserRepository(db.DB)
	userService := app.NewUserService(userRepo)
	userServer := grpc.NewUserServerImpl(userService)

	// gRPC 服务注册
	r, err := etcd.NewServiceRegistry([]string{etcdAddr})
	if err != nil {
		logger.Fatalln(err.Error())
	}
	if r == nil {
		logger.Fatalln("cant register service")
		return
	}

	// 注意：公钥管理逻辑需要根据实际情况调整
	// 此处暂时简化，实际应从安全位置加载
	var servicePublicKey []byte
	err = r.Register(serviceName, rpcAddr, servicePublicKey)
	if err != nil {
		logger.Fatalln(err.Error())
		return
	}
	defer func(r *etcd.ServiceRegistry) {
		err := r.Close()
		if err != nil {
			logger.Fatalln(err.Error())
		}
	}(r)

	// 启动 gRPC 服务
	server := grpc.NewServer()
	userPb.RegisterUserServer(server, userServer)
	lis, err := net.Listen("tcp", serviceAddr)
	if err != nil {
		logger.Fatalf("failed to listen:%v\n", err)
		return
	}

	fmt.Printf("user service listening at %v\n", serviceAddr)
	if err := server.Serve(lis); err != nil {
		logger.Fatalf("failed to serve:%v", err)
		return
	}
}
