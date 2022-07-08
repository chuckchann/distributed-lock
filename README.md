# distributed-lock

![avatar](https://img.shields.io/badge/build-unkown-orange)
![avatar](https://img.shields.io/badge/release-v1.0.0-brightgreen)
![avatar](https://img.shields.io/badge/license-unkown-yellow)

distributed-lock is a distributed mutex lock written in Go. It provides three implements, including redis, etcd and zookeeper(todo).

## Implement

- [x] redis
- [x] etcd
- [ ] zookeeper

------



## How to use 

download package

```shell
go get -u github.com/chuckchann/distributed-lock
```

example 

```go
func main()  {
	//init client
	redis_lock.Init(entry.Config{
		Endpoints:   []string{"127.0.0.1:6379"},
		DBIndex:     15,
		MaxConns:    20,
		IdleTimeout: 10 * time.Second,
		Password: 	"123456",
	})
	
	//create a new redis lock
	l := distributed_lock.New("test")
  
  //acquire lock
	err := l.Lock()
	if err != nil {
		return
	}

	//acquire lock successfully
	//handle logic business ...

	l.UnLock()
}
```




------

## TODO

- redis lock: change old lock policy(self spin if current lock was occupied by other client), the new lock policy just like [redission](https://github.com/redisson/redisson).
- zookeeper implement. 

