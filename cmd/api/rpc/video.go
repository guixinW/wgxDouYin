package rpc

import (
	"context"
	"fmt"
	"wgxDouYin/grpc/video"
	"wgxDouYin/pkg/viper"
)

var (
	videoClient video.VideoServiceClient
)

func InitVideo(config *viper.Config) {
	etcdAddresses := []string{fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))}
	serviceName := config.Viper.GetString("service.name")
	initGrpcClient(etcdAddresses, serviceName, &videoClient)
}

func Feed(ctx context.Context, req *video.FeedRequest) (*video.FeedResponse, error) {
	fmt.Println("call Feed")
	return videoClient.Feed(ctx, req)
}

func PublishAction(ctx context.Context, req *video.PublishActionRequest) (*video.PublishActionResponse, error) {
	fmt.Println("call PublishAction")
	return videoClient.PublishAction(ctx, req)
}

func PublishList(ctx context.Context, req *video.PublishListRequest) (*video.PublishListResponse, error) {
	fmt.Println("call PublishList")
	return videoClient.PublishList(ctx, req)
}
