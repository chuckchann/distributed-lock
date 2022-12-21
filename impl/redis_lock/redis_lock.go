package redis_lock

//ATTENTION!!!
//1. only suit for signal node redis server or cluster redis server
//2. redis server version >= 2.6.0

import (
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"github.com/google/uuid"
	"log"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultTTL     = 10 * time.Second
	DefaultTimeout = time.Second
	WatchInterval  = 5 * time.Second
)

type RedisLock struct {
	*entry.Options
	bizKey      string
	value       string
	stopWatchCh chan struct{}
}

func New(bizKey string, opts ...entry.Option) *RedisLock {
	o := &entry.Options{
		TTL:    DefaultTTL,
		NoSpin: false,
		Watch:  false,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.Logger == nil {
		o.Logger = log.Default()
	}
	var ch chan struct{}
	if o.Watch {
		ch = make(chan struct{}, 1) //buffered channel prevent goroutine leak
	}
	return &RedisLock{
		Options:     o,
		bizKey:      bizKey,
		stopWatchCh: ch,
	}
}

func (rl *RedisLock) TryLock() error {
	//generate redis key & value
	k := rl.generateRedisKey()
	v := rl.generateRedisValue()

	if rl.TTL <= 0 {
		rl.TTL = DefaultTTL
	}

	b, err := redisClient.SetNX(k, v, rl.TTL).Result()
	if err != nil {
		rl.Logger.Printf("RedisLock：TryLock [SetNX] failed %v \n", err)
		return err
	}
	if !b {
		return impl.ErrLockFailed
	}

	//get lock success
	rl.value = v

	if rl.Watch {
		go rl.watch()
	}

	return nil
}

func (rl *RedisLock) Lock() error {
	t := &time.Timer{}
	if rl.Timeout != 0 {
		t = time.NewTimer(rl.Timeout)
		defer t.Stop()
	}

	//generate redis key & value
	k := rl.generateRedisKey()
	v := rl.generateRedisValue()
	for {
		select {
		case <-t.C:
			return impl.ErrLockTimeout
		default:
			b, err := redisClient.SetNX(k, v, rl.TTL).Result()
			if err != nil {
				rl.Logger.Printf("RedisLock: Lock [SetNX] failed: %v \n", err)
				return err
			}
			if !b {
				//get lock failed
				if rl.NoSpin {
					//no spin: yield current p, allowing other g to obtain cpu
					runtime.Gosched()
				}
				break
			}

			rl.value = v

			//watch the key after acquire lock successfully
			if rl.Watch {
				go rl.watch()
			}

			return nil
		}
	}
}

func (rl *RedisLock) UnLock() error {
	if rl.value == "" {
		return impl.ErrUnLock
	}
	res, err := redisClient.Eval(UnlockScript, []string{rl.generateRedisKey()}, rl.value).Result()
	if err != nil {
		rl.Logger.Printf("RedisLock: UnLock [Eval] failed: %v \n ", err)
		return err
	}
	if res.(int64) == 0 {
		rl.Logger.Printf("RedisLock: bizKey %s has already expired \n", rl.bizKey)
	}

	//notify the watch goroutine to stop watch the key
	if rl.Watch {
		rl.stopWatchCh <- struct{}{}
	}

	return nil
}

func (rl *RedisLock) ForceKillScript() {
	redisClient.ScriptKill()
}

func (rl *RedisLock) generateRedisKey() string {
	if entry.GlobalPrefix != "" {
		return strings.Join([]string{entry.GlobalPrefix, "redis_lock", rl.bizKey}, ":")
	} else {
		return strings.Join([]string{"redis_lock", rl.bizKey}, ":")
	}
}

func (rl *RedisLock) generateRedisValue() string {
	return uuid.New().String()
}

func (rl *RedisLock) watch() {
	t := time.NewTicker(WatchInterval)
	key := rl.generateRedisKey()
	defer t.Stop()
	for {
		select {
		case <-t.C:
			//renewal key for 30s if key still exist
			res, err := redisClient.Eval(RenewalScript, []string{key}, 30000).Result()
			if err != nil {
				rl.Logger.Printf("RedisLock: [Eval] renewal key %s failed: %v \n", key, err)
			} else {
				if r, _ := res.(int64); r == 1 {
					rl.Logger.Printf("RedisLock: renewal key %s successfully \n", key)
				}
			}
		case <-rl.stopWatchCh:
			return
		}
	}
}
