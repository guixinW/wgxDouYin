package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"wgxDouYin/grpc/user"
	pb "wgxDouYin/grpc/user"
)

func main() {
	etcdClient, err := clientv3.NewFromURL("10.21.29.203:2379")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	etcdResolverBuilder, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	conn, err := grpc.NewClient(
		"etcd:///wgxDouYinUserServer",
		grpc.WithResolvers(etcdResolverBuilder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	greeter1 := pb.NewUserServiceClient(conn)
	req := &user.UserRegisterRequest{
		Username: "testUser",
		Password: "1477364283",
	}
	resp, err := greeter1.UserRegister(context.Background(), req)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	fmt.Printf("reply:%v\n", resp)
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(conn)
}
