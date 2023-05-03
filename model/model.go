package model

import (
	"context"
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db          *gorm.DB
	Rdb         *redis.Client
	conf        = config.Config
	modelLogger = log.Log
)

func init() {
	connectPostgreSQL()
	connectRedis()
}

func connectPostgreSQL() {
	var err error
	database := conf.Sub("postgres")
	username := database.GetString("username")
	password := database.GetString("password")
	databasename := database.GetString("dbname")
	host := database.GetString("host")
	port := database.GetInt("port")

	dsn := fmt.Sprintf(`host=%s user=%s
		password=%s dbname=%s
		port=%d sslmode=disable 
		TimeZone=Asia/Shanghai`,
		host,
		username,
		password,
		databasename,
		port,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		modelLogger.Panicln(err)
	}
}

func connectRedis() {
	redisConf := conf.Sub("redis")
	host := redisConf.GetString("host")
	port := redisConf.GetInt("port")
	Addr := host + ":" + fmt.Sprint(port)
	Password := redisConf.GetString("password")
	DB := redisConf.GetInt("db")
	Rdb = redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password,
		DB:       DB, // use default DB
	})
	modelLogger.Info("redis connect to %s, default DB is %s", Addr, DB)
	ctx := context.Background()
	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		modelLogger.Panicln("redis connect failed")
	}
}
