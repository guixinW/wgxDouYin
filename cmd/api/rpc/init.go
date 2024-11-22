package rpc

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"wgxDouYin/grpc/user"
	"wgxDouYin/pkg/viper"
)

func init() {
	userConfig := viper.Init("user")
	InitUser(&userConfig)
}

func initClient(scheme, serviceName string, client interface{}) {
	conn, err := connectServer(scheme, serviceName)
	if err != nil {
		panic(err)
	}
	switch c := client.(type) {
	case *user.UserServiceClient:
		*c = user.NewUserServiceClient(conn)
	default:
		panic("unsupported client type")
	}
}

func connectServer(scheme, serviceName string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := fmt.Sprintf("%s:///%s", scheme, serviceName)
	fmt.Printf("connect addr %v\n", addr)
	conn, err = grpc.NewClient(addr, opts...)
	return
}
