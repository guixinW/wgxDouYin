package etcd

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/resolver"
	"sync"
)

var (
	schema = "etcd"
)

type ServerToAddressMap struct {
	data sync.Map
}

func (p *ServerToAddressMap) Store(key string, value resolver.Address) {
	p.data.Store(key, value)
}

func (p *ServerToAddressMap) Load(key string) (resolver.Address, bool) {
	value, ok := p.data.Load(key)
	if !ok {
		return resolver.Address{}, false
	}
	return value.(resolver.Address), true
}

type TikTokServiceResolver struct {
	target  resolver.Target
	cc      resolver.ClientConn
	address ServerToAddressMap
}

func (r *TikTokServiceResolver) Update(key, value []byte) error {
	r.address.Store(string(key), resolver.Address{ServerName: string(key), Addr: string(value)})
	r.address.data.Range(func(k, v interface{}) bool {
		return true
	})
	if r.cc != nil {
		err := r.update()
		if err != nil {
			return err
		}
	}
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

func (r *TikTokServiceResolver) ResolveNow(options resolver.ResolveNowOptions) {
	err := r.update()
	if err != nil {
		return
	}
}

func (r *TikTokServiceResolver) Close() {}

func (r *TikTokServiceResolver) update() error {
	var updateAddress []resolver.Address
	r.address.data.Range(func(k, v interface{}) bool {
		if k.(string) == r.target.String()[8:] {
			updateAddress = append(updateAddress, v.(resolver.Address))
		}
		return true
	})
	err := r.cc.UpdateState(resolver.State{Addresses: updateAddress})
	if err != nil {
		return errors.Wrap(err, "TikTokServiceResolver update failed")
	}
	return nil
}
