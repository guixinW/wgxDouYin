package rpc

import (
	"context"
	"fmt"
	"wgxDouYin/grpc/user"
	"wgxDouYin/pkg/viper"
)

var (
	userClient user.UserServiceClient
)

func InitUser(config *viper.Config) {
	etcdAddresses := []string{fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))}
	serviceName := config.Viper.GetString("service.name")
	initGrpcClient(etcdAddresses, serviceName, &userClient)
}

func Register(ctx context.Context, req *user.UserRegisterRequest) (*user.UserRegisterResponse, error) {
	fmt.Println("call Register")
	return userClient.UserRegister(ctx, req)
}

func Login(ctx context.Context, req *user.UserLoginRequest) (*user.UserLoginResponse, error) {
	fmt.Println("call Login")
	return userClient.Login(ctx, req)
}

func UserInform(ctx context.Context, req *user.UserInfoRequest) (*user.UserInfoResponse, error) {
	fmt.Println("call UserInform")
	return userClient.UserInfo(ctx, req)
}

func AccessToken(ctx context.Context, req *user.AccessTokenRequest) (*user.AccessTokenResponse, error) {
	fmt.Println("call AccessToken")
	return userClient.RefreshAccessToken(ctx, req)
}
