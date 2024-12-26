package etcd

import (
	"context"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var (
	timeout      = time.Duration(5)
	syncInterval = time.Minute
)

type QueryUpdater interface {
	Update(key, value []byte) error
}

type QueryTool struct {
	cli                 *clientv3.Client
	closeChan           chan struct{}
	keyPrefixUpdaterMap map[string]QueryUpdater
}

func NewQueryTool(etcdAddress []string) (*QueryTool, error) {
	etcdCli, err := clientv3.New(clientv3.Config{Endpoints: etcdAddress})
	if err != nil {
		return nil, err
	}
	return &QueryTool{
		cli:                 etcdCli,
		closeChan:           make(chan struct{}),
		keyPrefixUpdaterMap: make(map[string]QueryUpdater),
	}, nil
}

func (tool *QueryTool) SetQuerySourceAddress(address []string) {
	tool.cli.SetEndpoints(address...)
}

func (tool *QueryTool) RegisterUpdater(keyPrefix string, updater QueryUpdater) {
	tool.keyPrefixUpdaterMap[keyPrefix] = updater
}

func (tool *QueryTool) notifyUpdater(serverName string, key, value []byte) error {
	err := tool.keyPrefixUpdaterMap[serverName].Update(key, value)
	if err != nil {
		return errors.Wrap(err, "notify updater failed")
	}
	return nil
}

func (tool *QueryTool) Watch(keyPrefix string) error {
	ticker := time.NewTicker(syncInterval)
	err := tool.sync(keyPrefix)
	if err != nil {
		return errors.Wrap(err, "Watch failed")
	}
	watchChan := tool.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			err := tool.update(keyPrefix, resp.Events)
			if err != nil {
				return errors.Wrap(err, "Watch failed")
			}
		case <-ticker.C:
			err := tool.sync(keyPrefix)
			if err != nil {
				return errors.Wrap(err, "Watch failed")
			}
		case <-tool.closeChan:
			return nil
		}
	}
}

func (tool *QueryTool) update(keyPrefix string, event []*clientv3.Event) error {
	for _, ev := range event {
		if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
			return tool.notifyUpdater(keyPrefix, ev.Kv.Key, ev.Kv.Value)
		}
	}
	return nil
}

func (tool *QueryTool) sync(keyPrefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	res, err := tool.cli.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return errors.Wrap(err, "sync failed")
	}
	for _, v := range res.Kvs {
		err = tool.notifyUpdater(keyPrefix, v.Key, v.Value)
		if err != nil {
			return errors.Wrap(err, "sync failed")
		}
	}
	return nil
}

func (tool *QueryTool) close() {
	tool.closeChan <- struct{}{}
}
