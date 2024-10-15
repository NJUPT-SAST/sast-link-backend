package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/NJUPT-SAST/sast-link-backend/config"
)

const (
	RedisKeyPrefix = "sast-link:"
)

// Store provides database access to all raw objects.
type Store struct {
	Profile *config.Config
	db      *gorm.DB
	rdb     *redis.Client
	log     *zap.Logger
}

type StoreOptions func(s *Store)

func (s *Store) WithOptions(opts ...StoreOptions) *Store {
	newStore := *s
	for _, opt := range opts {
		opt(&newStore)
	}
	return &newStore
}

func WithLogger(log *zap.Logger) StoreOptions {
	return func(s *Store) {
		s.log = log
	}
}

// WithLogger will copy store and set the logger to the new store.
func (s *Store) WithLogger(log *zap.Logger) *Store {
	newStore := *s
	newStore.log = log
	return &newStore
}

// NewStore creates a new store.
func NewStore(ctx context.Context, profile *config.Config, logger *zap.Logger) (*Store, error) {
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
		log:     logger,
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
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("Failed to connect database: %s\n", err)
		return nil, err
	}
	fmt.Printf("Connected to database: %s:%d\n", profile.PostgresHost, profile.PostgresPort)
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
		fmt.Printf("Failed to connect redis: %s\n", err)
		return nil, err
	}
	fmt.Printf("Connected to redis: %s:%d\n", host, port)
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
//
// If the value is a struct, it need to implement encoding.BinaryMarshaler.
func (s *Store) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	key = RedisKeyPrefix + key
	err := s.rdb.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get gets the value of a key.
//
// If the key does not exist, Get returns "".
func (s *Store) Get(ctx context.Context, key string) (string, error) {
	key = RedisKeyPrefix + key
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		// Return nil if the key does not exist
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

// Delete deletes a key.
func (s *Store) Delete(ctx context.Context, key string) error {
	key = RedisKeyPrefix + key
	err := s.rdb.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

// Exec executes a query without returning any rows.
func (s *Store) Exec(ctx context.Context, query string, args ...interface{}) error {
	tx := s.db.WithContext(ctx).Exec(query, args...)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (s *Store) SelectOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	tx := s.db.WithContext(ctx).Raw(query, args...).First(dest)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
