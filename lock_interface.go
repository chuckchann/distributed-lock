package distributed_lock

import (
	"distributed-lock/entry"
	"distributed-lock/impl"
	"distributed-lock/impl/etcd_lock"
	"distributed-lock/impl/redis_lock"
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
}

