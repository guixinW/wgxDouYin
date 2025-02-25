package rpc

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"wgxDouYin/grpc/relation"
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
	relationConfig := viper.Init("relation")
	InitUser(&userConfig)
	InitRelation(&relationConfig)
}

func errorHandler(err error, msg string) {
	if err != nil {
		logger.Errorln(errors.Wrap(err, msg))
	}
}

// initClient 初始化各类微服务的rpc客户端
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
		QueryTool.SetQuerySourceAddress(etcdAddress)
	}
	fmt.Printf("serviceName:%v\n", serviceName)
	conn, err := connectServer(serviceName)
	errorHandler(err, errMsg)
	switch c := client.(type) {
	case *user.UserServiceClient:
		*c = user.NewUserServiceClient(conn)
	case *relation.RelationServiceClient:
		*c = relation.NewRelationServiceClient(conn)
	default:
		panic("unsupported client type")
	}
}

// connectServer 连接到serviceName指定的rpc服务端
func connectServer(serviceName string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := fmt.Sprintf("%s:///%s", "etcd", etcd.AddrPrefix(serviceName))
	conn, err = grpc.NewClient(addr, opts...)
	return
}
