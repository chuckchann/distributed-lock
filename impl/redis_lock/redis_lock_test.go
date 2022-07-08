package redis_lock

import (
	"distributed-lock/entry"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

const host = "127.0.0.1:6379"

func TestRedisLock_Lock(t *testing.T) {
	Init(entry.Config{
		Endpoints:   []string{host},
		DBIndex:     15,
		MaxConns:    20,
		IdleTimeout: 10 * time.Second,
		Password: "123456",
	})

	m := make(map[int]bool, 100)
	iphone := 50
	rl := New("buy-iphone", entry.WithTTL(time.Second * 10))
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := rl.Lock(); err != nil {
				log.Println("获取锁失败")
				return
			}
			defer rl.UnLock()
			if iphone > 0 {
				fmt.Printf( "%d号用户 抢到iPhone 当前剩余iphone: %d \n", i, iphone)
				iphone--
				m[i] = true
			} else {
				fmt.Printf("%d号用户 没抢到 很遗憾\n", i)
				m[i] = false
			}
		}(i)

	}

	wg.Wait()
	var count1, count2 int
	for _, v := range m {
		if v {
			count1++
		} else {
			count2++
		}
	}

	fmt.Println("抢到 ", count1, "没抢到 ", count2)

}

func TestRedisLock_Lock2(t *testing.T) {
	Init(entry.Config{
		Endpoints:   []string{host},
		DBIndex:     15,
		MaxConns:    20,
		IdleTimeout: 10 * time.Second,
	})

	bizKey := "test"
	wg := &sync.WaitGroup{}
	wg.Add(2)


	go func() {
		defer wg.Done()
		rl := New(bizKey, entry.WithTTL(5 * time.Second))
		if err := rl.Lock(); err != nil {
			t.Fatal("1获取锁失败")
		}
		t.Log("1获取锁成功")

		t.Log("1 doing something...")
		time.Sleep(6 * time.Second)

		rl.UnLock()
	}()

	go func() {
		defer wg.Done()
		rl := New(bizKey, entry.WithTTL(5 * time.Second))
		if err := rl.Lock(); err != nil {
			t.Fatal("2获取锁失败")
		}
		t.Log("2获取锁成功")

		t.Log("2 doing something...")
		time.Sleep(6 * time.Second)

		rl.UnLock()
	}()

	wg.Wait()
	time.Sleep(2 * time.Second)
}

func Test_lua(t *testing.T)  {
	Init(entry.Config{
		Endpoints:   []string{host},
		DBIndex:     15,
		MaxConns:    20,
		IdleTimeout: 10 * time.Second,
	})
	res, err := redisClient.Eval(UnlockScript, []string{"test"}, "a").Result()
	if err != nil {
		t.Fatal("UnLock [Eval] failed: ", err)
	}
	t.Log(res, err)
}

func Test_lua2(t *testing.T)  {
	Init(entry.Config{
		Endpoints:   []string{host},
		DBIndex:     15,
		MaxConns:    20,
		IdleTimeout: 10 * time.Second,
		Password: "123456",
	})
	res, err := redisClient.Eval(RenewalScript, []string{"test2"}, 30000).Result()
	if err != nil {
		t.Fatal("UnLock [Eval] failed: ", err)
	}
	t.Log(res, err)
}

