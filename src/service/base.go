package service

import (
	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/store"
)

// BaseService is a base service
type BaseService struct {
	Store  *store.Store
	Config *config.Config
}

// NewBaseService creates a new base service
func NewBaseService(store *store.Store, config *config.Config) *BaseService {
    return &BaseService{
        Store:  store,
        Config: config,
    }
}
