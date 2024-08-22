package store

import (
	"github.com/pkg/errors"

	
	"github.com/NJUPT-SAST/sast-link-backend/plugin/idp/oauth2"
)

func (s *Store) CreateIdentityProvider(idp *oauth2.IdentityProviderSetting) error {
	// Check if the identity provider already exists
	if _, err := s.GetIdentityProviderByName(idp.Name); err == nil {
		return errors.Errorf("identity provider %s already exists", idp.Name)
	}
	return s.db.Table("idp").Create(idp).Error
}

func (s *Store) GetIdentityProviderByName(name string) (*oauth2.IdentityProviderSetting, error) {
	var idp oauth2.IdentityProviderSetting
	if err := s.db.Table("idp").Where("name = ?", name).First(&idp).Error; err != nil {
		return nil, err
	}
	return &idp, nil
}
