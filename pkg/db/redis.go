package db

import (
	"fmt"
	"sync"

	conf "github.com/iceymoss/go-hichat-api/pkg/config"

	"github.com/go-redis/redis/v8"
)

const HICHAT2_RDB = "main"

var redisConn = make(map[string]*redis.Client)
var redisMutex sync.RWMutex

func GetRedisConn() *redis.Client {
	redisMutex.Lock()
	rdb, ok := redisConn[HICHAT2_RDB]
	redisMutex.Unlock()
	if !ok {
		redisMutex.Lock()
		opt := redis.Options{
			Addr:     fmt.Sprintf("%s:%d", conf.ServiceConf.RedisDB.Host, conf.ServiceConf.RedisDB.Port),
			Password: conf.ServiceConf.RedisDB.PassWord,
			DB:       0,
		}
		rdb = redis.NewClient(&opt)
		redisConn[HICHAT2_RDB] = rdb
		redisMutex.Unlock()
	}
	return rdb
}
