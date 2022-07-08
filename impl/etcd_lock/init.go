package etcd_lock

import (
	"distributed-lock/entry"
	"distributed-lock/impl"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var etcdClinet *clientv3.Client

func Init(cfg entry.Config) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		panic("init ectd client failed, " + err.Error())
	}

	impl.UsingType = 2

	etcdClinet = c
}
