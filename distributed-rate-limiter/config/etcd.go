package config

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/SheoranRavi/drl/model"
	"github.com/SheoranRavi/drl/util"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Etcd struct {
	options EtcdOptions
	client  *clientv3.Client
}

func (e *Etcd) GetAllRules(ctx context.Context) map[string]model.Rule {
	// rules be like /ratelimit/rules/authorized/v1/profile
	rules := make(map[string]model.Rule)
	resp, err := e.client.Get(ctx, util.RULES_KEY, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		rule := model.Rule{}
		json.Unmarshal(kv.Value, &rule)
		key := string(kv.Key)
		rules[key] = rule
	}
	return rules
}

func (e *Etcd) Put(ctx context.Context, key string, value string) {
	e.client.Put(ctx, key, value)
}

func (e *Etcd) GetWatchCh(ctx context.Context) clientv3.WatchChan {
	watchCh := e.client.Watch(
		ctx,
		util.RULES_KEY,
		clientv3.WithPrefix(),
	)
	return watchCh
}

func (e *Etcd) connect() {
	var err error
	e.client, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{e.options.Address},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Etcd) Close() {
	e.client.Close()
}

func GetEtcd(options EtcdOptions) *Etcd {
	e := &Etcd{options: options}
	e.connect()
	return e
}

type EtcdOptions struct {
	Address string
}
