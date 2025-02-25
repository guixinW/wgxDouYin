package rpc

import (
	"context"
	"fmt"
	"wgxDouYin/grpc/relation"
	"wgxDouYin/pkg/viper"
)

var (
	relationClient relation.RelationServiceClient
)

func InitRelation(config *viper.Config) {
	etcdAddresses := []string{fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))}
	serviceName := config.Viper.GetString("service.name")
	initClient(etcdAddresses, serviceName, &relationClient)
}

func RelationAction(ctx context.Context, req *relation.RelationActionRequest) (*relation.RelationActionResponse, error) {
	fmt.Println("call relation action")
	return relationClient.RelationAction(ctx, req)
}

//func RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (*relation.RelationFollowListResponse, error) {
//	return relationClient.RelationFollowList(ctx, req)
//}
//
//func RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (*relation.RelationFollowerListResponse, error) {
//	return relationClient.RelationFollowerList(ctx, req)
//}
//
//func RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (*relation.RelationFriendListResponse, error) {
//	return relationClient.RelationFriendList(ctx, req)
//}
