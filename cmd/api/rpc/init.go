package rpc

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"wgxDouYin/grpc/user"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/viper"
)

func init() {
	userConfig := viper.Init("user")
	InitUser(&userConfig)
}

// initClient this function connect to rpc server by service name. service name will be
// resolved by resolver that is init by InitUser.
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

// connectServer is used by initClient.
func connectServer(scheme, serviceName string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := fmt.Sprintf("%s:///%s", scheme, etcd.AddrPrefix(serviceName))
	fmt.Printf("connect addr %v\n", addr)
	conn, err = grpc.NewClient(addr, opts...)
	return
}

func initPublicKey(scheme, serviceName string) {

}
