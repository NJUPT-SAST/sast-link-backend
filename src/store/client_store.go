package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/pkg/errors"

	"github.com/NJUPT-SAST/sast-link-backend/util"
)

type ClientStore struct {
	dbStore   Store
	tableName string
}

// ClientStoreItem data item
//
// Distinct from the model in oauth2 package, this struct is used to
// store data in the database and the model in oauth2 package is used
// in oauth2 process.
type ClientStoreItem struct {
	ID     string `gorm:"id"`
	Secret string `gorm:"secret"`
	Domain string `gorm:"domain"`
	Data   []byte `gorm:"data"` // Data store OAuth2 pacakage client model data
	UserID string `gorm:"user_id"`
	Name   string `goem:"name"` // Name is the client name
	Desc   string `gorm:"desc"` // Desc is the client description
}

func (c *ClientStoreItem) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

func (c *ClientStoreItem) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

type FindClientRequest struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type UpdateClientRequest struct {
	ID           string   `json:"id"`
	UserID       string   `json:"user_id"`
	Name         string   `json:"name"`
	Desc         string   `json:"desc"`
	RedirectURIs []string `json:"redirect_uri"` // Domain
	// Other fields ...
}

// NewClientStore creates PostgreSQL store instance.
func NewClientStore(dbStore Store) *ClientStore {
	store := &ClientStore{
		dbStore:   dbStore,
		tableName: "oauth2_clients",
	}

	return store
}

// nolint
// TODO: All tables should be created in the same place.
func (s *ClientStore) initTable() error {
	// Create table if not exists
	if err := s.dbStore.db.Migrator().CreateTable(&ClientStoreItem{}); err != nil {
		return err
	}
	return s.dbStore.Exec(context.Background(), fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %[1]s (
  id      TEXT  NOT NULL,
  secret  TEXT  NOT NULL,
  domain  TEXT  NOT NULL,
  data    JSONB NOT NULL,
  name    TEXT  NOT NULL,
  user_id TEXT NOT NULL,
  desc    TEXT,
  CONSTRAINT %[1]s_pkey PRIMARY KEY (id)
);
`, s.tableName))
}

func (*ClientStore) toClientInfo(data []byte) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal(data, &cm)
	return &cm, err
}

// GetByID retrieves and returns client information by id
//
// GetByID is implemented to satisfy the ClientStore interface in
// the oauth2 package.
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

// ListClient retrieves and returns client information by user id.

// It reveives a FindClientRequest struct as a parameter, which
// contains the user id and client id.
func (s *ClientStore) ListClient(ctx context.Context, find FindClientRequest) ([]ClientStoreItem, error) {
	// Initialize the result slice
	result := make([]ClientStoreItem, 0)

	query := s.dbStore.db.Table(s.tableName).WithContext(ctx)
	var err error

	if find.UserID == "" && find.ID == "" {
		return nil, errors.New("invalid request: UserID or ID must be provided")
	}

	// Determine query based on find conditions
	if find.ID != "" {
		query = query.Where("id = ?", find.ID)
	}
	if find.UserID != "" {
		query = query.Where("user_id = ?", find.UserID)
	}

	// Execute the query
	if err = query.Find(&result).Error; err != nil {
		return nil, err
	}

	// Process the client secrets
	for i := range result {
		result[i].Secret = util.MaskSecret(result[i].Secret)
	}

	return result, nil
}

// GetClient retrieves client, but it will return all client information.
func (s *ClientStore) GetClient(ctx context.Context, find FindClientRequest) (*ClientStoreItem, error) {
	if find.ID == "" {
		return nil, errors.New("invalid request: ID must be provided")
	}

	if cache, err := s.dbStore.Get(ctx, find.ID); err == nil && cache != "" {
		var item ClientStoreItem
		if err := json.Unmarshal([]byte(cache), &item); err == nil {
			return &item, nil
		}
	}

	list, err := s.ListClient(ctx, find)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, nil
	}

	client := list[0]
	// cache the client info(30 days)
	_ = s.dbStore.Set(ctx, client.ID, client, 30*24*60*60)
	return &client, nil
}

// UpdateClient updates the client information.
func (s *ClientStore) UpdateClient(ctx context.Context, request UpdateClientRequest) error {
	tx := s.dbStore.db.Table(s.tableName).WithContext(ctx).Begin()

	if request.ID == "" {
		return errors.New("invalid request: ID must be provided")
	}
	if request.UserID == "" {
		return errors.New("invalid request: UserID must be provided")
	}

	updateData := map[string]interface{}{}
	if request.Name != "" {
		updateData["name"] = request.Name
	}
	if request.Desc != "" {
		updateData["desc"] = request.Desc
	}
	if len(request.RedirectURIs) > 0 {
		updateData["domain"] = strings.Join(request.RedirectURIs, ",")
	}

	if err := tx.Where("id = ?", request.ID).
		Where("user_id = ?", request.UserID).
		Updates(updateData).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *ClientStore) DeleteClient(ctx context.Context, id, uid string) error {
	return s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", id).Where("user_id = ?", uid).Delete(&ClientStoreItem{}).Error
}

// Create creates and stores the new client information.
func (s *ClientStore) Create(ctx context.Context, info oauth2.ClientInfo, name, desc string) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return s.dbStore.db.Table(s.tableName).WithContext(ctx).Create(&ClientStoreItem{
		ID:     info.GetID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		Data:   data,
		UserID: info.GetUserID(),
		Name:   name,
		Desc:   desc,
	}).Error
}

// AddRedirectURI adds redirect uri to client information.
func (s *ClientStore) AddRedirectURI(ctx context.Context, uid, clientID, uri string) error {
	if clientID == "" || uri == "" {
		return nil
	}

	var item ClientStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", clientID).First(&item).Error; err != nil {
		return err
	}

	dbMap := make(map[string]interface{})
	if err := json.Unmarshal(item.Data, &dbMap); err != nil {
		return err
	}

	dbURI := item.Domain
	dbUid, ok := dbMap["UserID"].(string)
	if !ok {
		return errors.New("user id not found")
	}

	if dbUid != uid {
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

	return s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", clientID).Update("domain", strings.Join(uris, ",")).Error
}

// CheckPermissions checks the user whether has the permission to access the client.
func (s *ClientStore) CheckClientOwner(ctx context.Context, uid, clientID string) bool {
	var item ClientStoreItem
	if err := s.dbStore.db.Table(s.tableName).WithContext(ctx).Where("id = ?", clientID).First(&item).Error; err != nil {
		return false
	}

	dbMap := make(map[string]interface{})
	if err := json.Unmarshal(item.Data, &dbMap); err != nil {
		return false
	}

	dbUid, ok := dbMap["UserID"].(string)
	if !ok {
		return false
	}

	if dbUid != uid {
		return false
	}

	return true
}
