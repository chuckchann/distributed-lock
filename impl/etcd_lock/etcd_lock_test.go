package etcd_lock

import (
	"distributed-lock/entry"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestEtcdLock_Lock(t *testing.T) {
	Init(entry.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})

	iphone := 60
	wg := &sync.WaitGroup{}
	m := make(map[int]bool, 100)

	for i := 0; i < 100; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()

			l := New("/test")
			if err := l.Lock(); err != nil {
				fmt.Println("get lock err ", err)
				return
			}
			defer l.Unlock()

			//模拟耗时操作
			time.Sleep(5 * time.Second)

			if iphone >= 1 {
				iphone--
				fmt.Println(i, "抢到了iphone 剩余iphone数量 ", iphone)
				m[i] = true
			} else {
				fmt.Println(i, " 很遗憾，抢完了")
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

	fmt.Println(count1, count2)

}
