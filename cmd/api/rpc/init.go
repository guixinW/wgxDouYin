package rpc

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"wgxDouYin/grpc/relation"
	"wgxDouYin/grpc/user"
	"wgxDouYin/grpc/video"
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
	if err != nil {
		logger.Errorln(errors.Wrap(err, "init error"))
	}
	TikTokResolver = &etcd.TikTokServiceResolver{}
	resolver.Register(TikTokResolver)
	userConfig := viper.Init("user")
	relationConfig := viper.Init("relation")
	videoConfig := viper.Init("video")
	InitUser(&userConfig)
	InitRelation(&relationConfig)
	InitVideo(&videoConfig)
}

func errorHandler(err error, msg string) {
	if err != nil {
		logger.Errorln(errors.Wrap(err, msg))
	}
}

// initGrpcClient 初始化各类微服务的rpc客户端
func initGrpcClient(etcdAddress []string, serviceName string, client interface{}) {
	initDiscovery(etcdAddress, serviceName)
	initClient(serviceName, client)
}

func initClient(serviceName string, client interface{}) {
	conn, err := connectServer(serviceName)
	if err != nil {
		errorHandler(err, "connect server error")
	}
	switch c := client.(type) {
	case *user.UserServiceClient:
		*c = user.NewUserServiceClient(conn)
	case *relation.RelationServiceClient:
		*c = relation.NewRelationServiceClient(conn)
	case *video.VideoServiceClient:
		*c = video.NewVideoServiceClient(conn)
	default:
		panic("unsupported client type")
	}
}

// 初始化etcd服务发现实例，并注册需要发现的服务名、密钥名
func initDiscovery(etcdAddress []string, serviceName string) {
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
	}
}

// connectServer 连接到serviceName指定的rpc服务端
func connectServer(serviceName string) (conn *grpc.ClientConn, err error) {
	addr := fmt.Sprintf("%s:///%s", "etcd", etcd.AddrPrefix(serviceName))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err = grpc.NewClient(addr, opts...)
	if err != nil {
		fmt.Printf("connect server error: %v\n", err)
	}
	return
}
