package zookeeper_lock

import (
	"errors"
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"github.com/go-zookeeper/zk"
	"log"
)

type ZookeeperLock struct {
	*entry.Options
	bizKey string
	lock   *zk.Lock
}

func New(bizKey string, opts ...entry.Option) *ZookeeperLock {
	o := &entry.Options{}
	for _, opt := range opts {
		opt(o)
	}
	if o.Logger == nil {
		o.Logger = log.Default()
	}
	return &ZookeeperLock{
		bizKey: bizKey,
	}
}

func (zl *ZookeeperLock) Lock() error {
	lock := zk.NewLock(zookeeperClient, zl.bizKey, nil)
	err := lock.Lock()
	if err != nil {
		zl.Logger.Printf("ZookeeperLock: Lock [Lock] failed  %v \n", err)
		return err
	}
	zl.lock = lock
	return nil
}

func (*ZookeeperLock) TryLock() error {
	return errors.New("zookeeper don`t support TryLock API")
}

func (zl *ZookeeperLock) UnLock() error {
	if zl.lock == nil {
		return impl.ErrUnLock
	}

	return zl.lock.Unlock()
}
