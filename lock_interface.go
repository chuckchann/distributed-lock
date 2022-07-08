package distributed_lock

import (
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"github.com/chuckchann/distributed-lock/impl/etcd_lock"
	"github.com/chuckchann/distributed-lock/impl/redis_lock"
)

type DistributedLock interface {
	Lock() error
	TryLock() error
	UnLock() error
}

//init client before you use lock!!
func New(bizKey string, opts ...entry.Option) DistributedLock {
	switch impl.UsingType {
	case 1:
		return redis_lock.New(bizKey, opts...)
	case 2:
		return etcd_lock.New(bizKey, opts...)
	case 3:
		//return zookeeper_lock.New(bizKey, opts...)
	default:
		return redis_lock.New(bizKey, opts...)
	}
	return nil
}

