package redis_lock

//ATTENTION!!!
//1. only suit for signal node redis server or cluster redis server
//2. redis server version >= 2.6.0

import (
	"github.com/chuckchann/distributed-lock/entry"
	"github.com/chuckchann/distributed-lock/impl"
	"github.com/google/uuid"
	"gopkg.in/redis.v5"
	"log"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultTTL           = 10 * time.Second
	RetryLockInterval    = 100 * time.Millisecond
	DefaultMaxRetryTimes = 30
	ChannelMessage       = "lock has been released"
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
		NoSpin: true,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.Logger == nil {
		o.Logger = log.Default()
	}
	if o.TTL <= 0 {
		o.TTL = DefaultTTL
	}
	if o.MaxRetryTimes <= 0 {
		o.MaxRetryTimes = DefaultMaxRetryTimes
	}
	var ch chan struct{}
	if o.RenewalTime > 0 {
		ch = make(chan struct{}, 1) //buffered channel prevent goroutine leak
	}
	return &RedisLock{
		Options:     o,
		bizKey:      bizKey,
		stopWatchCh: ch,
	}
}


func (rl *RedisLock) generateRedisKey() string {
	if entry.GlobalPrefix != "" {
		return strings.Join([]string{entry.GlobalPrefix, "redis_lock", rl.bizKey}, ":")
	} else {
		return strings.Join([]string{"redis_lock", rl.bizKey}, ":")
	}
}

func (rl *RedisLock) generateChannelKey() string {
	return rl.generateRedisKey() + ":lock_channel"
}

func (rl *RedisLock) generateRedisValue() string {
	return uuid.New().String()
}

func (rl *RedisLock) tryLock() (acquire bool, err error) {
	k := rl.generateRedisKey()
	v := rl.generateRedisValue()
	acquire, err = redisClient.SetNX(k, v, rl.TTL).Result()
	if err != nil {
		rl.Logger.Printf("RedisLock: Lock [SetNX] failed: %v \n", err)
		return
	}
	// acquire lock successfully
	if acquire {
		rl.value = v
		// watch the key after acquire lock successfully
		// inspired by https://github.com/redisson/redisson
		if rl.RenewalTime > 0 {
			go rl.watch()
		}
	}
	return
}

func (rl *RedisLock) TryLock() error {
	// try to acquire lock
	acquired, err := rl.tryLock()
	if err != nil {
		return err
	}
	if acquired {
		return nil
	}
	return impl.ErrLockFailed
}

func (rl *RedisLock) Lock() error {
	t := &time.Timer{}
	if rl.Timeout != 0 {
		t = time.NewTimer(rl.Timeout)
		defer t.Stop()
	}


	retryTimes := 0
	addRetryTimes := func() {
		retryTimes++
	}

	var sub *redis.PubSub
	var leftTimeout time.Duration
	start := time.Now()

	for {
		if retryTimes < rl.MaxRetryTimes { //retry mode
			select {
			case <-t.C:
				return impl.ErrLockTimeout
			default:
				// try to acquire lock
				acquired, err := rl.tryLock()
				if err != nil {
					return err
				}
				if acquired {
					// acquire lock successfully
					return nil
				}

				// acquire lock failed
				if rl.NoSpin {
					// no spin: yield current p, allowing other goroutine to obtain cpu
					runtime.Gosched()
				}
				addRetryTimes()

				if retryTimes <= rl.MaxRetryTimes {
					time.Sleep(RetryLockInterval)
				} else {
					// otherwise retry times over MaxRetryTimes, don`t retry "SetNX" any more, case we don`t want to interactive
					// redis server frequently since we have already try MaxRetryTimes (which may case a lot of network IO). instead
					// we subscribe a channel and turn to awake mode. when we receive message from this channel, that means this
					// client can "SetNX" again.
					var err error
					sub, err = redisClient.Subscribe(rl.generateChannelKey())
					if err != nil {
						rl.Logger.Printf("RedisLock: Lock [Subscribe] failed: %v \n", err)
						return err
					}
					if rl.Timeout != 0 {
						// calculate left Timeout
						leftTimeout =   rl.Timeout - time.Now().Sub(start)
					}
					rl.Logger.Printf("RedisLock: Lock ready enter awake mode")
				}
				break
			}
		} else { //awake mode. inspired by https://github.com/redisson/redisson
			var err error
			// block and wait for unlock message
			if rl.Timeout != 0 {
				_, err = sub.ReceiveTimeout(leftTimeout)
			}else {
				_, err = sub.ReceiveMessage()
			}
			if err != nil {
				rl.Logger.Printf("RedisLock: Lock [ReceiveMessage] failed: %v \n", err)
				return err
			}
			// try to acquire lock
			acquired, err := rl.tryLock()
			if err != nil {
				return err
			}
			if acquired {
				// unsubscribe message channel
				if err := sub.Unsubscribe(rl.generateChannelKey()); err != nil {
					rl.Logger.Printf("RedisLock: Lock [Unsubscribe] failed: %v \n", err)
					return err
				}
				return nil
			}
		}
	}
}

func (rl *RedisLock) UnLock() error {
	if rl.value == "" {
		return impl.ErrUnLock
	}
	res, err := redisClient.Eval(UnlockScript, []string{rl.generateRedisKey(), rl.generateChannelKey()}, rl.value, ChannelMessage).Result()
	if err != nil {
		rl.Logger.Printf("RedisLock: UnLock [Eval] failed: %v \n ", err)
		return err
	}
	if res.(int64) == 1 {
		rl.Logger.Printf("RedisLock: bizKey %s has already expired \n", rl.bizKey)
	}

	// notify the watch goroutine to stop watch the key
	if rl.RenewalTime > 0 {
		rl.stopWatchCh <- struct{}{}
	}

	return nil
}

func (rl *RedisLock) watch() {
	// watch interval: rrl/3
	t := time.NewTicker(rl.TTL / 3)
	key := rl.generateRedisKey()
	defer t.Stop()
	for {
		select {
		case <-t.C:
			// renewal key for 30s if key still exist
			res, err := redisClient.Eval(RenewalScript, []string{key}, rl.value, rl.RenewalTime.Milliseconds()).Result()
			if err != nil {
				rl.Logger.Printf("RedisLock: [Eval] renewal key %s failed: %v \n", key, err)
			} else {
				r, _ := res.(int64)
				// 0: still exist
				// 1: key has expired
				// 2: key has been occupied by another client
				if r == 0 || r == 1 {
					rl.Logger.Printf("RedisLock: renewal key %s successfully \n", key)
				} else {
					rl.Logger.Printf("RedisLock: key %s is illegal now !!! \n", key)
					return //finish watch
				}
			}
		case <-rl.stopWatchCh:
			return //finish watch
		}
	}
}

func (rl *RedisLock) ForceKillScript() {
	redisClient.ScriptKill()
}
