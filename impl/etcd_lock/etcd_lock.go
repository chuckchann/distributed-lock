package etcd_lock

import (
	"context"
	"distributed-lock/entry"
	"errors"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
)

type EtcdLock struct {
	*entry.OptionConfig
	bizKey string
	value  string
	m      *concurrency.Mutex
	s      *concurrency.Session
}

func New(bizKey string, opts ...entry.Option) *EtcdLock {
	return &EtcdLock{
		bizKey: bizKey,
	}
}

func (el *EtcdLock) TryLock() error {
	s, err := concurrency.NewSession(etcdClinet)
	if err != nil {
		log.Println("etcd new session failed, ", err.Error())
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
		log.Println("etcd new session failed, ", err.Error())
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
		str := "etcd session or mutex is nil"
		log.Println(str)
		return errors.New(str)
	}
	defer el.s.Close()

	return el.m.Unlock(context.Background())
}
