package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	REDIS_KEY_PREFIX = "sast-link:"
)

// Store provides database access to all raw objects.
type Store struct {
	Profile *config.Config
	db      *gorm.DB
	rdb     *redis.Client
}

// NewStore creates a new store.
func NewStore(ctx context.Context, profile *config.Config) (*Store, error) {
	db, err := NewPostgresDB(profile)
	if err != nil {
		return nil, err
	}
	rdb, err := NewRedisDB(ctx, profile)
	if err != nil {
		return nil, err
	}
	return &Store{
		Profile: profile,
		db:      db,
		rdb:     rdb,
	}, nil
}

func NewPostgresDB(profile *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(`host=%s user=%s
    		password=%s dbname=%s
    		port=%d sslmode=disable 
    		TimeZone=Asia/Shanghai`,
		profile.PostgresHost,
		profile.PostgresUser,
		profile.PostgresPWD,
		profile.PostgresDB,
		profile.PostgresPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Panicf("Failed to connect database: %s", err)
		return nil, err
	}
	log.Infof("Connected to database: %s:%d", profile.PostgresHost, profile.PostgresPort)
	return db, nil
}

func NewRedisDB(ctx context.Context, profile *config.Config) (*redis.Client, error) {
	host := profile.RedisHost
	port := profile.RedisPort
	addr := fmt.Sprintf("%s:%d", host, port)
	// maxIdle := redisConf.GetInt("maxIdle")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: profile.RedisPWD,
		DB:       profile.RedisDB, // use default DB
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil || rdb == nil {
		log.Panicf("Failed to connect redis: %s", err)
		return nil, err
	}
	log.Infof("Connected to redis: %s", addr)
	return rdb, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	var errs []string

	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			errs = append(errs, fmt.Sprintf("retrieve sql.DB: %s", err))
		} else if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close sql database: %s", err))
		}
	}

	if s.rdb != nil {
		if err := s.rdb.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close redis: %s", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close store resources: %s", strings.Join(errs, "; "))
	}

	return nil
}

// Set sets a key-value pair with expiration time.
func (s *Store) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	key = REDIS_KEY_PREFIX + key
	err := s.rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get gets the value of a key.
func (s *Store) Get(ctx context.Context, key string) (string, error) {
	key = REDIS_KEY_PREFIX + key
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// Delete deletes a key.
func (s *Store) Delete(ctx context.Context, key string) error {
	key = REDIS_KEY_PREFIX + key
	err := s.rdb.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
