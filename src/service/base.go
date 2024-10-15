package service

import (
	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"go.uber.org/zap"
)

// BaseService is a base service.
type BaseService struct {
	Store  *store.Store
	Config *config.Config

	log *zap.Logger
}

// NewBaseService creates a new base service.
func NewBaseService(store *store.Store, config *config.Config) *BaseService {
	return &BaseService{
		Store:  store,
		Config: config,
	}
}

type BaseSvcOptions func(s *BaseService)

func (s *BaseService) WithOptions(opts ...BaseSvcOptions) *BaseService {
	newService := *s
	for _, opt := range opts {
		opt(&newService)
	}
	return &newService
}

func WithLogger(log *zap.Logger) BaseSvcOptions {
	return func(s *BaseService) {
		s.log = log
	}
}

func WithStore(store *store.Store) BaseSvcOptions {
	return func(s *BaseService) {
		s.Store = store
	}
}
