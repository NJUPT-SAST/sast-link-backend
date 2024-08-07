package model

import (
	"context"
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	redisSession "github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	Db          *gorm.DB
	Rdb         *redis.Client
	conf        = config.Config
	modelLogger = log.Log
	redisConf   = GetRedisConf()
	postgreConf = GetPostgresConf()
	RedisCtx    = context.Background()
)

// Redis config
type RedisConf struct {
	Host     string
	Port     int
	Addr     string
	Password string
	Db       int
	MaxIdle  int
}

// Postgres config
type PostgresConf struct {
	Host     string
	Port     int
	Username string
	Password string
	Dbname   string
}

func init() {
	connectPostgreSQL()
	connectRedis()
	// newStore()
}

// Get redis config
func GetRedisConf() *RedisConf {
	redisConf := conf.Sub("redis")
	host := redisConf.GetString("host")
	port := redisConf.GetInt("port")
	addr := fmt.Sprintf("%s:%d", host, port)
	password := redisConf.GetString("password")
	db := redisConf.GetInt("db")
	maxIdle := redisConf.GetInt("maxIdle")
	return &RedisConf{
		Host:     host,
		Port:     port,
		Addr:     addr,
		Password: password,
		Db:       db,
		MaxIdle:  maxIdle,
	}
}

// Get postgres config
func GetPostgresConf() *PostgresConf {
	database := conf.Sub("postgres")
	host := database.GetString("host")
	port := database.GetInt("port")
	username := database.GetString("username")
	password := database.GetString("password")
	dbname := database.GetString("dbname")
	return &PostgresConf{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Dbname:   dbname,
	}
}

func newStore() {
	addr := redisConf.Addr
	db := redisConf.Db
	store := redisSession.NewRedisStore(&redisSession.Options{
		Addr: addr,
		DB:   db,
	})
	session.InitManager(session.SetStore(store))
}

func connectPostgreSQL() {
	var err error
	username := postgreConf.Username
	password := postgreConf.Password
	databasename := postgreConf.Dbname
	host := postgreConf.Host
	port := postgreConf.Port

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

	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		modelLogger.Panicln(err)
	}
}

func connectRedis() {
	Addr := redisConf.Addr
	Password := redisConf.Password
	DB := redisConf.Db
	Rdb = redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password,
		DB:       DB, // use default DB
	})

	_, err := Rdb.Ping(RedisCtx).Result()
	if err != nil || Rdb == nil {
		modelLogger.Panicln("redis connect failed")

	}
	modelLogger.Infof("Redis connect to [Addr: %s], [DB: %d]\n", Addr, DB)
}
