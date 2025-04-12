package rpc

import (
	"context"
	"fmt"
	"wgxDouYin/grpc/comment"
	"wgxDouYin/pkg/viper"
)

var (
	commentClient comment.CommentServiceClient
)

func InitComment(config *viper.Config) {
	etcdAddresses := []string{fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))}
	serviceName := config.Viper.GetString("service.name")
	initGrpcClient(etcdAddresses, serviceName, &commentClient)
}

func CommentAction(ctx context.Context, req *comment.CommentActionRequest) (*comment.CommentActionResponse, error) {
	fmt.Println("call favorite action")
	return commentClient.CommentAction(ctx, req)
}

func CommentList(ctx context.Context, req *comment.CommentListRequest) (*comment.CommentListResponse, error) {
	fmt.Println("call favorite list")
	return commentClient.CommentList(ctx, req)
}
