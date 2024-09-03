package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"gorm.io/gorm"

	"github.com/NJUPT-SAST/sast-link-backend/log"
)

// TokenStore store oauth2 token information
type TokenStore struct {
	dbStore   Store
	tableName string

	gcDisabled bool
	gcInterval time.Duration
	ticker     *time.Ticker
}

// TokenStoreItem data item
type TokenStoreItem struct {
	ID        int64     `gorm:"id"`
	CreatedAt time.Time `gorm:"created_at"`
	ExpiresAt time.Time `gorm:"expires_at"`
	Code      string    `gorm:"code"`
	Access    string    `gorm:"access"`
	Refresh   string    `gorm:"refresh"`
	Data      []byte    `gorm:"data"`
}

// NewTokenStore create token store instance
func NewTokenStore(dbStore Store, options ...TokenStoreOption) *TokenStore {
	store := &TokenStore{
		dbStore:    dbStore,
		tableName:  "oauth2_tokens",
		gcInterval: 10 * time.Minute,
	}

	for _, o := range options {
		o(store)
	}

	if !store.gcDisabled {
		store.ticker = time.NewTicker(store.gcInterval)
		go store.gc()
	}

	return store
}

// Close closes the store
func (s *TokenStore) Close() error {
	if !s.gcDisabled {
		s.ticker.Stop()
	}
	return nil
}

func (s *TokenStore) gc() {
	for range s.ticker.C {
		s.clean()
	}
}

// TODO: other tables should be created, so we need to create a new function to create tables
func (s *TokenStore) initTable() error {
	return s.dbStore.Exec(context.Background(), fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %[1]s (
	id         BIGSERIAL   NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	code       TEXT        NOT NULL,
	access     TEXT        NOT NULL,
	refresh    TEXT        NOT NULL,
	data       JSONB       NOT NULL,
	CONSTRAINT %[1]s_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_%[1]s_expires_at ON %[1]s (expires_at);
CREATE INDEX IF NOT EXISTS idx_%[1]s_code ON %[1]s (code);
CREATE INDEX IF NOT EXISTS idx_%[1]s_access ON %[1]s (access);
CREATE INDEX IF NOT EXISTS idx_%[1]s_refresh ON %[1]s (refresh);
`, s.tableName))
}

func (s *TokenStore) clean() {
	now := time.Now()
	if err := s.dbStore.db.Table(s.tableName).Where("expires_at <= ?", now).Delete(&TokenStoreItem{}).Error; err != nil {
		log.Errorf("Error while cleaning out outdated entities: %+v", err)
	}
}

// Create creates and stores the new token information
func (s *TokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	buf, err := json.Marshal(info)
	if err != nil {
		return err
	}

	item := &TokenStoreItem{
		Data:      buf,
		CreatedAt: time.Now(),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiresAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())

		// Cache the code
		s.dbStore.Set(ctx, code, buf, info.GetCodeExpiresIn())
	} else {
		item.Access = info.GetAccess()
		item.ExpiresAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())

		s.dbStore.Set(ctx, info.GetAccess(), buf, info.GetAccessExpiresIn())

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiresAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())

			s.dbStore.Set(ctx, refresh, buf, info.GetRefreshExpiresIn())
		}
	}

	return s.dbStore.db.Table(s.tableName).WithContext(ctx).Create(item).Error
}

// RemoveByCode deletes the authorization code
func (s *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	s.dbStore.Delete(ctx, code)
	err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("code = ?", code).Delete(&TokenStoreItem{}).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	return err
}

// RemoveByAccess uses the access token to delete the token information
func (s *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	s.dbStore.Delete(ctx, access)
	err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("access = ?", access).Delete(&TokenStoreItem{}).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	return err
}

// RemoveByRefresh uses the refresh token to delete the token information
func (s *TokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	s.dbStore.Delete(ctx, refresh)
	err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("refresh = ?", refresh).Delete(&TokenStoreItem{}).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	return err
}

func (s *TokenStore) toTokenInfo(data []byte) (oauth2.TokenInfo, error) {
	var tm models.Token
	err := json.Unmarshal(data, &tm)
	return &tm, err
}

// GetByCode uses the authorization code for token information data
func (s *TokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	if data, err := s.dbStore.Get(ctx, code); err == nil {
		return s.toTokenInfo([]byte(data))
	}

	var item TokenStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).First(&item).Where("code = ?", code).Error; err != nil {
		return nil, err
	}

	return s.toTokenInfo(item.Data)
}

// GetByAccess uses the access token for token information data
func (s *TokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	if data, err := s.dbStore.Get(ctx, access); err == nil {
		return s.toTokenInfo([]byte(data))
	}

	var item TokenStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).First(&item).Where("access = ?", access).Error; err != nil {
		return nil, err
	}

	return s.toTokenInfo(item.Data)
}

// GetByRefresh uses the refresh token for token information data
func (s *TokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	if data, err := s.dbStore.Get(ctx, refresh); err == nil {
		return s.toTokenInfo([]byte(data))
	}

	var item TokenStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).First(&item).Where("refresh = ?", refresh).Error; err != nil {
		return nil, err
	}

	return s.toTokenInfo(item.Data)
}

// TokenStoreOption is the configuration options type for token store
type TokenStoreOption func(s *TokenStore)

// WithTokenStoreGCInterval returns option that sets token store garbage collection interval
func WithTokenStoreGCInterval(gcInterval time.Duration) TokenStoreOption {
	return func(s *TokenStore) {
		s.gcInterval = gcInterval
	}
}

// WithTokenStoreGCDisabled returns option that disables token store garbage collection
func WithTokenStoreGCDisabled() TokenStoreOption {
	return func(s *TokenStore) {
		s.gcDisabled = true
	}
}
