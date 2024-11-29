package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"wgxDouYin/cmd/user/service"
	userPb "wgxDouYin/grpc/user"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	config      = viper.Init("user")
	serviceName = config.Viper.GetString("server.name")
	serviceAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("server.host"),
		config.Viper.GetInt("server.port"))
	rpcAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("rpc.host"),
		config.Viper.GetInt("rpc.port"))
	etcdAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"),
		config.Viper.GetInt("etcd.port"))
	logger = zap.InitLogger()
)

func init() {
	privateKeyPath := fmt.Sprintf("keys/%v.pem", serviceName)
	privateKey, err := keys.LoadPrivateKey(privateKeyPath)
	if err != nil {
		panic(err)
	}
	service.Init(privateKey, serviceName)
}

func main() {
	r, err := etcd.NewServiceRegistry([]string{etcdAddr})
	if err != nil {
		logger.Fatalln(err.Error())
	}
	if r == nil {
		logger.Fatalln("cant register service")
		return
	}

	servicePublicKey, err := service.KeyManager.GetServerPublicKey(serviceName)
	if err != nil || servicePublicKey == nil {
		logger.Fatalln("cant get service public key")
	}
	err = r.Register(serviceName, rpcAddr, servicePublicKey)
	if err != nil {
		logger.Fatalln(err.Error())
	}
	defer func(r *etcd.ServiceRegistry) {
		err := r.Close()
		if err != nil {
			logger.Fatalln(err.Error())
		}
	}(r)
	server := grpc.NewServer()
	userPb.RegisterUserServiceServer(server, &service.UserServerImpl{})
	lis, err := net.Listen("tcp", serviceAddr)
	fmt.Printf("listen %v\n", serviceAddr)
	if err != nil {
		logger.Fatalf("failed to listen:%v\n", err)
	}
	if err := server.Serve(lis); err != nil {
		logger.Fatalf("failed to serve:%v", err)
	}
}
