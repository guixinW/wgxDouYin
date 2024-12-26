package etcd

import (
	"context"
	"crypto/ecdsa"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"strconv"
	"time"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/zap"
)

var (
	ttlKey     = ""
	defaultTTL = 60
	ctxTimeout = 3
)

type registerMeta struct {
	leaseID clientv3.LeaseID
	ctx     context.Context
	cancel  context.CancelFunc
}

type ServiceRegistry struct {
	etcdClient *clientv3.Client
	leaseTTL   int64
	meta       *registerMeta
}

func NewServiceRegistry(endpoints []string) (*ServiceRegistry, error) {
	return NewServiceRegistryWithAuth(endpoints, "", "")
}

func NewServiceRegistryWithAuth(endpoints []string, username, password string) (*ServiceRegistry, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, err
	}
	return &ServiceRegistry{
		etcdClient: etcdClient,
		leaseTTL:   getTTL(),
	}, nil
}

func (e *ServiceRegistry) Register(serviceName, serviceAddr string, servicePublicKey *ecdsa.PublicKey) error {
	leaseID, err := e.createLease()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(ctxTimeout))
	defer cancel()

	//将service的地址存入etcd
	if serviceName != "" && serviceAddr != "" {
		_, err = e.etcdClient.Put(ctx, AddrPrefix(serviceName),
			serviceAddr, clientv3.WithLease(leaseID))
		if err != nil {
			return err
		}
	}

	//将service的公钥存入etcd
	if servicePublicKey.Curve != nil {
		servicePublicKeyString, err := keys.PublicKeyToPEM(servicePublicKey)
		if err != nil {
			return err
		}
		_, err = e.etcdClient.Put(ctx, KeyPrefix(serviceName),
			servicePublicKeyString, clientv3.WithLease(leaseID))
		if err != nil {
			return err
		}
	}

	meta := registerMeta{
		leaseID: leaseID,
	}
	meta.ctx, meta.cancel = context.WithCancel(context.Background())
	e.meta = &meta
	if err := e.keepAlive(); err != nil {
		return err
	}
	return nil
}

func (e *ServiceRegistry) createLease() (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(ctxTimeout))
	defer cancel()
	resp, err := e.etcdClient.Grant(ctx, e.leaseTTL)
	if err != nil {
		return clientv3.NoLease, err
	}
	return resp.ID, nil
}

func (e *ServiceRegistry) keepAlive() error {
	logger := zap.InitLogger()
	keepAliveChan, err := e.etcdClient.KeepAlive(context.Background(), e.meta.leaseID)
	if err != nil {
		return err
	}
	go func(keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse) {
		logger.Infof("start keepalive lease %x for etcd register", e.meta.leaseID)
		for range keepAliveChan {
			select {
			case <-e.meta.ctx.Done():
				break
			default:
			}
		}
		logger.Infof("stop keepalive lease %x for etcd register", e.meta.leaseID)
	}(keepAliveChan)
	return nil
}

func (e *ServiceRegistry) Close() error {
	_, err := e.etcdClient.Revoke(context.Background(), e.meta.leaseID)
	if err != nil {
		return err
	}
	return e.etcdClient.Close()
}

func getTTL() int64 {
	var ttl int64 = int64(defaultTTL)
	if str, ok := os.LookupEnv(ttlKey); ok {
		if t, err := strconv.Atoi(str); err == nil {
			ttl = int64(t)
		}
	}
	return ttl
}
