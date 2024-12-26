package etcd

import (
	"github.com/pkg/errors"
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

func (r *TikTokServiceResolver) Update(key, value []byte) error {
	if r.address == nil {
		r.address = make([]resolver.Address, 0)
	}
	r.address = append(r.address, resolver.Address{ServerName: string(key), Addr: string(value)})
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
	err := r.cc.UpdateState(resolver.State{Addresses: r.address})
	if err != nil {
		return errors.Wrap(err, "TikTokServiceResolver update failed")
	}
	return nil
}
