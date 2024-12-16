package rpc

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"wgxDouYin/grpc/user"
	"wgxDouYin/pkg/etcd"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	logger         = zap.InitLogger()
	QueryTool      *etcd.QueryTool
	KeysManager    *keys.KeyManager
	TikTokResolver *etcd.TikTokServiceResolver
)

func init() {
	var err error
	KeysManager, err = keys.NewKeyManager(nil, "")
	TikTokResolver = &etcd.TikTokServiceResolver{}
	resolver.Register(TikTokResolver)
	if err != nil {
		logger.Errorln(err.Error())
	}
	userConfig := viper.Init("user")
	InitUser(&userConfig)
}

// initClient this function connect to rpc server by service name. service name will be
// resolved by resolver that is init by InitUser.
func initClient(etcdAddress []string, serviceName string, client interface{}) {
	if QueryTool == nil {
		var err error
		QueryTool, err = etcd.NewQueryTool(etcdAddress)
		if err != nil {
			logger.Errorln(err.Error())
		}
		QueryTool.RegisterUpdater(etcd.KeyPrefix(serviceName), KeysManager)
		QueryTool.RegisterUpdater(etcd.AddrPrefix(serviceName), TikTokResolver)
		go QueryTool.Watch(etcd.KeyPrefix(serviceName))
		go QueryTool.Watch(etcd.AddrPrefix(serviceName))
	} else {
		QueryTool.SetQuerySourceAddress(etcdAddress)
	}
	conn, err := connectServer(serviceName)
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
func connectServer(serviceName string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := fmt.Sprintf("%s:///%s", "etcd", etcd.AddrPrefix(serviceName))
	fmt.Printf("connect addr %v\n", addr)
	conn, err = grpc.NewClient(addr, opts...)
	return
}
