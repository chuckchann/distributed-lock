# distributed-lock

![avatar](https://img.shields.io/badge/build-unkown-orange)
![avatar](https://img.shields.io/badge/release-v1.0.0-brightgreen)
![avatar](https://img.shields.io/badge/license-unkown-yellow)

distributed-lock is a high performance distributed mutex written in Go. It provides three implements, including redis, etcd and zookeeper.

## Implement

- [x] redis: guarantee **AP**
- [x] etcd: guarantee **CP**
- [x] zookeeper: guarantee **CP**

## Features

- provide multi implements, including redis, etcd and zookeeper.
- in redis implement: <u>use lua script to lock/unlock redis lock</u>, <u>watch lock after acquiring lock</u>, <u>subscribe a topic when acquire lock failed ,when receive message from this topic then can acquire lock again (which can reduce interaction with redis)</u>.   

## How to use

download package

```shell
go get -u github.com/chuckchann/distributed-lock
```

example 

```go
func main()  {
	// init redis/etcd/zookeeper client berfore you use lock !!!
	redis_lock.Init(entry.Config{
		Endpoints:   []string{"127.0.0.1:6379"},
		DBIndex:     15,
		MaxConns:    20,
		IdleTimeout: 10 * time.Second,
		Password: 	"123456",
	})
	
	// create a new redis lock
	l := distributed_lock.New("test")
	
    // acquire lock
	err := l.Lock()
	if err != nil {
		return
	}

	// acquire lock successfully

	// do something ...
	
	l.UnLock()
}
```


------

## TODO

- [x] redis lock: change old lock policy(self spin if current lock was occupied by other client), the new lock policy just like [redission](https://github.com/redisson/redisson).



