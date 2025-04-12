package main

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
	"wgxDouYin/cmd/comment/service"
	commentPb "wgxDouYin/grpc/comment"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	config      = viper.Init("comment")
	serviceName = config.Viper.GetString("service.name")
	serviceAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("service.host"),
		config.Viper.GetInt("service.port"))
	rpcAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("rpc.host"),
		config.Viper.GetInt("rpc.port"))
	etcdAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"),
		config.Viper.GetInt("etcd.port"))
	logger = zap.InitLogger()
)

func errorHandler(err error, errMsg string) {
	if err != nil {
		logger.Errorln(errors.Wrap(err, errMsg))
	}
}

func init() {
	errMsg := "init failed"
	privateKeyPath := fmt.Sprintf("keys/%v.pem", serviceName)
	privateKey, err := keys.LoadPrivateKey(privateKeyPath)
	errorHandler(err, errMsg)
	err = service.Init(privateKey, serviceName)
	errorHandler(err, errMsg)
}

func main() {
	errMsg := "relation service failed"
	r, err := etcd.NewServiceRegistry([]string{etcdAddr})
	errorHandler(err, errMsg)
	if r == nil {
		logger.Fatalln("cant register service")
		return
	}
	servicePublicKey, err := service.KeyManager.GetServerPublicKey(serviceName)
	errorHandler(err, errMsg)
	fmt.Printf("service publick key:%v, service name:%v\n", servicePublicKey, serviceName)
	err = r.Register(serviceName, rpcAddr, servicePublicKey)
	errorHandler(err, errMsg)
	server := grpc.NewServer()
	commentPb.RegisterCommentServiceServer(server, &service.CommentServiceImpl{})
	lis, err := net.Listen("tcp", serviceAddr)
	errorHandler(err, errMsg)
	fmt.Printf("listen %v\n", serviceAddr)
	err = server.Serve(lis)
	errorHandler(err, errMsg)
}
