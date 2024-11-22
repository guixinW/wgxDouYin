package rpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/resolver"
	"wgxDouYin/grpc/user"
	grpc "wgxDouYin/grpc/user"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/viper"
)

var (
	userClient grpc.UserServiceClient
)

func InitUser(config *viper.Config) {
	etcdAddr := fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))
	serviceName := config.Viper.GetString("server.name")
	Discoverer := etcd.NewDiscoverer([]string{etcdAddr})

	resolver.Register(Discoverer)
	//defer Discoverer.Close()
	initClient(Discoverer.Scheme(), serviceName, &userClient)
}

func Register(ctx context.Context, req *user.UserRegisterRequest) (*user.UserRegisterResponse, error) {
	return userClient.UserRegister(ctx, req)
}
