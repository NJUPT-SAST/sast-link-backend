package util

import (
	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/redis/go-redis/v9"
)

var (
	conf = config.Config
)

// 连接redis并返回redis操作变量
func ConnectRedis() *redis.Client {
	redisConf := conf.Sub("redis.localhost")
	Addr := redisConf.GetString("addr")
	Password := redisConf.GetString("password")
	DB := redisConf.GetInt("db")
	rdb := redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password,
		DB:       DB,
	})
	return rdb
}
