package redis_lock

//ATTENTION!!!
//1. only suit for signal node redis server or cluster redis server
//2. redis server version >= 2.6.0

import (
	"errors"
	"github.com/chuckchann/distributed-lock/entry"
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

var (
	ErrLockFailed  = errors.New("get lock failed")
	ErrLockTimeout = errors.New("get lock timeout")
	ErrEmptyVal    = errors.New("lock value is empty")
)


type RedisLock struct {
	*entry.OptionConfig
	bizKey      string
	value       string
	stopWatchCh chan struct{}
}

func New(bizKey string, opts ...entry.Option) *RedisLock {
	oc := &entry.OptionConfig{
		TTL:    DefaultTTL,
		NoSpin: false,
	}
	for _, opt := range opts {
		opt(oc)
	}
	return &RedisLock{
		OptionConfig: oc,
		bizKey:       bizKey,
		stopWatchCh:  make(chan struct{}, 1),  //prevent goroutine leak
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
		log.Println("TryLock [SetNX] failed:", err)
		return err
	}
	if !b {
		return ErrLockFailed
	}

	//get lock success
	rl.value = v

	go rl.watch()

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
			return ErrLockTimeout
		default:
			b, err := redisClient.SetNX(k, v, rl.TTL).Result()
			if err != nil {
				log.Println("Lock [SetNX] failed:", err)
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
			go rl.watch()

			return nil
		}
	}
}

func (rl *RedisLock) UnLock() error {
	if rl.value == "" {
		return ErrEmptyVal
	}
	res, err := redisClient.Eval(UnlockScript, []string{rl.generateRedisKey()}, rl.value).Result()
	if err != nil {
		log.Println("UnLock [Eval] failed: ", err.Error())
		return err
	}
	if res.(int64) == 0 {
		log.Printf("bizKey: %s has already expire ", rl.bizKey)
	}

	//notify the watch goroutine to stop watch the key
	rl.stopWatchCh <-struct {}{}

	return nil
}


func (rl *RedisLock) ForceKillScript() {
	redisClient.ScriptKill()
}

func (rl *RedisLock) generateRedisKey() string {
	if entry.GlobalPrefix != "" {
		return strings.Join([]string{entry.GlobalPrefix, "redis_lock", rl.bizKey}, ":")
	}else {
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
				log.Printf("renewal key %s failed: %s", key, err.Error())
			}else {
				if r, _ := res.(int64); r == 1 {
					log.Printf("renewal key: %s successfuly ", key)
				}
			}
		case <-rl.stopWatchCh:
			return
		}
	}
}
