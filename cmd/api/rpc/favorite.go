package rpc

import (
	"context"
	"fmt"
	"wgxDouYin/grpc/favorite"
	"wgxDouYin/pkg/viper"
)

var (
	favoriteClient favorite.FavoriteServiceClient
)

func InitFavorite(config *viper.Config) {
	etcdAddresses := []string{fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))}
	serviceName := config.Viper.GetString("service.name")
	initGrpcClient(etcdAddresses, serviceName, &favoriteClient)
}

func FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (*favorite.FavoriteActionResponse, error) {
	fmt.Println("call favorite action")
	return favoriteClient.FavoriteAction(ctx, req)
}

func FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (*favorite.FavoriteListResponse, error) {
	fmt.Println("call favorite list")
	return favoriteClient.FavoriteList(ctx, req)
}
