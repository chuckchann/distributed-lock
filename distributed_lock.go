package distributed_lock

import (
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"github.com/chuckchann/distributed-lock/impl/etcd_lock"
	"github.com/chuckchann/distributed-lock/impl/redis_lock"
	"github.com/chuckchann/distributed-lock/impl/zookeeper_lock"
)

type DistributedLock interface {
	Lock() error
	TryLock() error
	UnLock() error
}

//init client before you use lock!!
func New(bizKey string, opts ...entry.Option) DistributedLock {
	switch impl.Type {
	case impl.REDIS:
		return redis_lock.New(bizKey, opts...)
	case impl.ETCD:
		return etcd_lock.New(bizKey, opts...)
	case impl.ZOOKEEPER:
		return zookeeper_lock.New(bizKey, opts...)
	default:
		return redis_lock.New(bizKey, opts...)
	}
}

//suggest set your project name as GlobalPrefix
func SetGlobalPrefix(p string) {
	entry.GlobalPrefix = p
}
