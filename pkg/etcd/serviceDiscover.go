package etcd

import (
	"fmt"
	"google.golang.org/grpc/resolver"
)

var (
	schema = "etcd"
)

type TikTokServiceResolver struct {
	target  resolver.Target
	cc      resolver.ClientConn
	address []resolver.Address
}

func (r *TikTokServiceResolver) Update(key, value []byte) {
	if r.address == nil {
		r.address = make([]resolver.Address, 0)
	}
	r.address = append(r.address, resolver.Address{ServerName: string(key), Addr: string(value)})
}

func (r *TikTokServiceResolver) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r.target = target
	r.cc = cc
	go r.update()
	return r, nil
}

func (r *TikTokServiceResolver) Scheme() string {
	return schema
}

func (r *TikTokServiceResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func (r *TikTokServiceResolver) Close() {}

func (r *TikTokServiceResolver) update() {
	fmt.Printf("resolver update:%v\n", r.address)
	err := r.cc.UpdateState(resolver.State{Addresses: r.address})
	if err != nil {
		zapLogger.Errorln(err.Error())
	}
}
