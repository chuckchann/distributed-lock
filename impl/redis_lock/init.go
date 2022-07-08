package redis_lock

import (
	"distributed-lock/entry"
	"distributed-lock/impl"
	"gopkg.in/redis.v5"
)

var redisClient *redis.Client

func Init(cfg entry.Config) {
	if len(cfg.Endpoints) == 0 {
		panic("endpoints is empty")
	}
	cli := redis.NewClient(&redis.Options{
		Addr:        cfg.Endpoints[0], //默认取第一个
		Password:    cfg.Password,
		DB:          cfg.DBIndex,
		MaxRetries:  5,
		IdleTimeout: cfg.IdleTimeout,
		PoolSize:    cfg.MaxConns,
	})

	_, err := cli.Ping().Result()
	if err != nil {
		panic("can not connect to redis, err:" + err.Error())
	}

	impl.UsingType = 1

	redisClient = cli
}
