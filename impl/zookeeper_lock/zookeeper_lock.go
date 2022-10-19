package zookeeper_lock

import (
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/go-zookeeper/zk"
)

type ZookeeperLock struct {
	*entry.OptionConfig
	bizKey string
	lock   *zk.Lock
}

func New(bizKey string, opts ...entry.Option) *ZookeeperLock {
	return &ZookeeperLock{
		bizKey: bizKey,
	}
}

func (zl *ZookeeperLock) Lock() error {
	lock := zk.NewLock(zookeeperClient, zl.bizKey, nil)
	err := lock.Lock()
	if err != nil {
		return err
	}
	return nil
}

func (*ZookeeperLock) TryLock() error {
	//zookeeper don`t support TryLock API

	return nil
}

func (zl *ZookeeperLock) UnLock() error {
	zookeeperClient.Set()
	return zl.lock.Unlock()
}
