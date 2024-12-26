package rpc

import (
	"fmt"
	"github.com/pkg/errors"
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
		logger.Errorln(errors.Wrap(err, "init error"))
	}
	userConfig := viper.Init("user")
	InitUser(&userConfig)
}

func errorHandler(err error, msg string) {
	if err != nil {
		logger.Errorln(errors.Wrap(err, msg))
	}
}

// initClient this function connect to rpc service by service name. service name will be
// resolved by resolver that is init by InitUser.
func initClient(etcdAddress []string, serviceName string, client interface{}) {
	var err error
	errMsg := "initClient failed"

	if QueryTool == nil {
		QueryTool, err = etcd.NewQueryTool(etcdAddress)
		errorHandler(err, errMsg)
		QueryTool.RegisterUpdater(etcd.KeyPrefix(serviceName), KeysManager)
		QueryTool.RegisterUpdater(etcd.AddrPrefix(serviceName), TikTokResolver)
		go func() {
			err := QueryTool.Watch(etcd.KeyPrefix(serviceName))
			errorHandler(err, errMsg)
		}()
		go func() {
			err := QueryTool.Watch(etcd.AddrPrefix(serviceName))
			errorHandler(err, errMsg)
		}()
	} else {
		QueryTool.SetQuerySourceAddress(etcdAddress)
	}

	conn, err := connectServer(serviceName)
	errorHandler(err, errMsg)
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
	conn, err = grpc.NewClient(addr, opts...)
	return
}
