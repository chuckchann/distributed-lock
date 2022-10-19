package zookeeper_lock

import (
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"github.com/go-zookeeper/zk"
	"time"
)

var zookeeperClient *zk.Conn

func Init(cfg entry.Config) {
	cli, _, err := zk.Connect(cfg.Endpoints, time.Second*5)
	if err != nil {
		panic("can not connect to zookeeper, err:" + err.Error())
	}
	zookeeperClient = cli

	impl.Type = impl.ZOOKEEPER
}
