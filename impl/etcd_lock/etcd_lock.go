package etcd_lock

import (
	"context"
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
)

type EtcdLock struct {
	*entry.Options
	bizKey string
	value  string
	m      *concurrency.Mutex
	s      *concurrency.Session
}

func New(bizKey string, opts ...entry.Option) *EtcdLock {
	o := &entry.Options{}
	for _, opt := range opts {
		opt(o)
	}
	if o.Logger == nil {
		o.Logger = log.Default()
	}
	return &EtcdLock{
		bizKey: bizKey,
	}
}

func (el *EtcdLock) TryLock() error {
	s, err := concurrency.NewSession(etcdClinet)
	if err != nil {
		el.Logger.Printf("EtcdLock: TryLock [NewSession] failed %v \n", err)
		return err
	}
	m := concurrency.NewMutex(s, el.bizKey)
	err = m.TryLock(context.TODO())
	if err != nil {
		return err
	}

	//store m and s
	el.m = m
	el.s = s

	return nil
}

func (el *EtcdLock) Lock() error {
	s, err := concurrency.NewSession(etcdClinet)
	if err != nil {
		el.Logger.Printf("EtcdLock: Lock [NewSession] failed %v \n", err)
		return err
	}
	m := concurrency.NewMutex(s, el.bizKey)
	if el.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), el.Timeout)
		defer cancel()
		err = m.Lock(ctx)
	} else {
		err = m.Lock(context.Background())
	}

	if err != nil {
		return err
	}

	//store m and s
	el.m = m
	el.s = s

	return nil
}

func (el *EtcdLock) UnLock() error {
	if el.m == nil || el.s == nil {
		return impl.ErrUnLock
	}
	defer el.s.Close()

	return el.m.Unlock(context.Background())
}
