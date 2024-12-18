package etcd

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
	"wgxDouYin/pkg/zap"
)

var (
	timeout      = time.Duration(5)
	zapLogger    = zap.InitLogger()
	syncInterval = time.Minute
)

type QueryUpdater interface {
	Update(key, value []byte)
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

func (tool *QueryTool) notifyUpdater(serverName string, key, value []byte) {
	tool.keyPrefixUpdaterMap[serverName].Update(key, value)
}

func (tool *QueryTool) Watch(keyPrefix string) {
	ticker := time.NewTicker(syncInterval)
	tool.sync(keyPrefix)
	watchChan := tool.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			tool.update(keyPrefix, resp.Events)
		case <-ticker.C:
			tool.sync(keyPrefix)
		case <-tool.closeChan:
			return
		}
	}
}

func (tool *QueryTool) update(keyPrefix string, event []*clientv3.Event) {
	for _, ev := range event {
		if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
			tool.notifyUpdater(keyPrefix, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

func (tool *QueryTool) sync(keyPrefix string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	res, err := tool.cli.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		zapLogger.Errorln(err.Error())
	}
	if res != nil {
		for _, v := range res.Kvs {
			tool.notifyUpdater(keyPrefix, v.Key, v.Value)
		}
	}
}

func (tool *QueryTool) close() {
	tool.closeChan <- struct{}{}
}
