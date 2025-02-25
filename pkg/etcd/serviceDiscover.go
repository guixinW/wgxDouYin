package etcd

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/resolver"
)

var (
	schema = "etcd"
)

type TikTokServiceResolver struct {
	target  resolver.Target
	cc      resolver.ClientConn
	address map[string]resolver.Address
}

func (r *TikTokServiceResolver) Update(key, value []byte) error {
	if r.address == nil {
		r.address = make(map[string]resolver.Address)
	}
	r.address[string(key)] = resolver.Address{ServerName: string(key), Addr: string(value)}
	fmt.Println(r.address)
	return nil
}

func (r *TikTokServiceResolver) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r.target = target
	r.cc = cc
	err := r.update()
	if err != nil {
		return nil, errors.Wrap(err, "Build Resolver Failed")
	}
	return r, nil
}

func (r *TikTokServiceResolver) Scheme() string {
	return schema
}

func (r *TikTokServiceResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func (r *TikTokServiceResolver) Close() {}

func (r *TikTokServiceResolver) update() error {
	var updateAddress []resolver.Address
	for _, value := range r.address {
		updateAddress = append(updateAddress, value)
	}
	err := r.cc.UpdateState(resolver.State{Addresses: updateAddress})
	if err != nil {
		return errors.Wrap(err, "TikTokServiceResolver update failed")
	}
	return nil
}
