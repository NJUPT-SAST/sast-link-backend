package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/pkg/errors"
)

// ClientStore PostgreSQL client store
type ClientStore struct {
	dbStore   Store
	tableName string
}

// ClientStoreItem data item
type ClientStoreItem struct {
	ID     string `gorm:"id"`
	Secret string `gorm:"secret"`
	Domain string `gorm:"domain"`
	Data   []byte `gorm:"data"`
}

// NewClientStore creates PostgreSQL store instance
func NewClientStore(dbStore Store) *ClientStore {
	store := &ClientStore{
		dbStore:   dbStore,
		tableName: "oauth2_clients",
	}

	return store
}

func (s *ClientStore) initTable() error {
	// Create table if not exists
	s.dbStore.db.Migrator().CreateTable(&ClientStoreItem{})
	return s.dbStore.Exec(context.Background(), fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %[1]s (
  id     TEXT  NOT NULL,
  secret TEXT  NOT NULL,
  domain TEXT  NOT NULL,
  data   JSONB NOT NULL,
  CONSTRAINT %[1]s_pkey PRIMARY KEY (id)
);
`, s.tableName))
}

func (s *ClientStore) toClientInfo(data []byte) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal(data, &cm)
	return &cm, err
}

// GetByID retrieves and returns client information by id
func (s *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}

	return s.toClientInfo(item.Data)
}

// Create creates and stores the new client information
func (s *ClientStore) Create(ctx context.Context, info oauth2.ClientInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return s.dbStore.db.Table(s.tableName).WithContext(ctx).Create(&ClientStoreItem{
		ID:     info.GetID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		Data:   data,
	}).Error
}

func (s *ClientStore) AddRedirectURI(ctx context.Context, stuID, id, uri string) error {
	if id == "" || uri == "" {
		return nil
	}

	var item ClientStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		return err
	}

	dbMap := make(map[string]interface{})
	if err := json.Unmarshal(item.Data, &dbMap); err != nil {
		return err
	}

	dbURI := item.Domain
	dbStuID := dbMap["UserID"].(string)

	if dbStuID != stuID {
		return errors.New("user id not match")
	}

	uris := strings.Split(dbURI, ",")
	if len(uris) > 0 {
		for _, u := range uris {
			if u == uri {
				return nil
			}
		}
	}

	uris = append(uris, uri)

	return s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", id).Update("domain", strings.Join(uris, ",")).Error
}
