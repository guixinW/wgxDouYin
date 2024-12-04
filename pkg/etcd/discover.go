package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"time"
	"wgxDouYin/pkg/zap"
)

var (
	timeout   = time.Duration(5)
	schema    = "etcd"
	zapLogger = zap.InitLogger()
)

type TikTokServiceResolverBuilder struct {
	EtcdAddresses []string
}

func (builder *TikTokServiceResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   builder.EtcdAddresses,
		DialTimeout: timeout * time.Second,
	})
	if err != nil {
		return nil, err
	}
	r := &TikTokServiceResolver{
		target:    target,
		cc:        cc,
		cli:       etcdClient,
		keyPrefix: target.URL.Path,
	}
	return r, nil
}

func (builder *TikTokServiceResolverBuilder) Scheme() string {
	return schema
}

type TikTokServiceResolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	cli       *clientv3.Client
	keyPrefix string
}

func (r *TikTokServiceResolver) ResolveNow(options resolver.ResolveNowOptions) {
	fmt.Println("resolver now")
	err := r.fetchAddressesFromEtcd()
	if err != nil {
		zapLogger.Fatalln(err)
		return
	}
}

func (r *TikTokServiceResolver) Close() {
	if r.cli != nil {
		err := r.cli.Close()
		if err != nil {
			zapLogger.Fatalln(err)
			return
		}
	}
}

func (r *TikTokServiceResolver) fetchAddressesFromEtcd() error {
	resp, err := r.cli.Get(context.Background(), r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	var addresses []resolver.Address
	for _, v := range resp.Kvs {
		addr := resolver.Address{ServerName: string(v.Key), Addr: string(v.Value)}
		addresses = append(addresses, addr)
	}
	fmt.Printf("service srvAddsList:%v\n", addresses)
	err = r.cc.UpdateState(resolver.State{Addresses: addresses})
	if err != nil {
		return err
	}
	return nil
}

//type ServiceDiscoverer struct {
//	schema   string
//	EtcdAdds []string
//
//	closeCh     chan struct{}
//	watchCh     clientv3.WatchChan
//	etcdClient  *clientv3.Client
//	keyPrefix   string
//	srvAddsList []resolver.Address
//
//	cc resolver.ClientConn
//}

//func NewDiscoverer(etcdAdds []string) *ServiceDiscoverer {
//	return &ServiceDiscoverer{
//		schema:   schema,
//		EtcdAdds: etcdAdds,
//	}
//}
//
//func (s *ServiceDiscoverer) Scheme() string {
//	return s.schema
//}
//
//func (s *ServiceDiscoverer) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
//	s.cc = cc
//	s.keyPrefix = target.Endpoint()
//	if _, err := s.start(); err != nil {
//		return nil, err
//	}
//	return s, nil
//}
//
//func (s *ServiceDiscoverer) ResolveNow(o resolver.ResolveNowOptions) {}
//
//func (s *ServiceDiscoverer) Close() {
//	s.closeCh <- struct{}{}
//}

//func (s *ServiceDiscoverer) start() (chan<- struct{}, error) {
//	var err error
//	s.etcdClient, _ = clientv3.New(clientv3.Config{
//		Endpoints:   s.EtcdAdds,
//		DialTimeout: timeout * time.Second,
//	})
//	if err != nil {
//		return nil, err
//	}
//	resolver.Register(s)
//
//	s.closeCh = make(chan struct{})
//
//	if err = s.sync(); err != nil {
//		return nil, err
//	}
//	fmt.Printf("srvAddsList:%v\n", s.srvAddsList)
//	go s.watch()
//
//	return s.closeCh, nil
//}
//
//func (s *ServiceDiscoverer) watch() {
//	ticker := time.NewTicker(time.Minute)
//	s.watchCh = s.etcdClient.Watch(context.Background(), s.keyPrefix, clientv3.WithPrefix())
//	for {
//		select {
//		case <-s.closeCh:
//			return
//		case res, ok := <-s.watchCh:
//			if ok {
//				if err := s.update(res.Events); err != nil {
//					panic(err)
//				}
//			}
//		case <-ticker.C:
//			if err := s.sync(); err != nil {
//				panic(err)
//			}
//		}
//	}
//}
//
//func (s *ServiceDiscoverer) update(events []*clientv3.Event) error {
//	for _, ev := range events {
//		switch ev.Type {
//		case mvccpb.PUT:
//			addr := resolver.Address{Addr: string(ev.Kv.Value)}
//			flag := false
//			for i := 0; i < len(s.srvAddsList); i++ {
//				if s.srvAddsList[i] == addr {
//					flag = true
//				}
//			}
//			if !flag {
//				s.srvAddsList = append(s.srvAddsList, addr)
//				err := s.cc.UpdateState(resolver.State{Addresses: s.srvAddsList})
//				if err != nil {
//					return err
//				}
//			}
//		case mvccpb.DELETE:
//			addr := resolver.Address{Addr: string(ev.Kv.Value)}
//			i := 0
//			for ; i < len(s.srvAddsList); i++ {
//				if s.srvAddsList[i] == addr {
//					break
//				}
//			}
//			if i < len(s.srvAddsList) {
//				s.srvAddsList = append(s.srvAddsList[:i], s.srvAddsList[:i+1]...)
//				err := s.cc.UpdateState(resolver.State{Addresses: s.srvAddsList})
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//	return nil
//}
//
//func (s *ServiceDiscoverer) sync() error {
//	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
//	defer cancel()
//	res, err := s.etcdClient.Get(ctx, s.keyPrefix, clientv3.WithPrefix())
//	if err != nil {
//		return err
//	}
//	s.srvAddsList = []resolver.Address{}
//	for _, v := range res.Kvs {
//		addr := resolver.Address{Addr: string(v.Value), ServerName: string(v.Key)}
//		s.srvAddsList = append(s.srvAddsList, addr)
//	}
//	fmt.Printf("service srvAddsList:%v\n", s.srvAddsList)
//	err = s.cc.UpdateState(resolver.State{Addresses: s.srvAddsList})
//	if err != nil {
//		return err
//	}
//	return nil
//}
