package store

import (
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"

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

func (s *Store) ListIdentityProviders() ([]oauth2.IdentityProviderSetting, error) {
	var idps []oauth2.IdentityProviderSetting
	if err := s.db.Table("idp").Find(&idps).Error; err != nil {
		return nil, err
	}
	return idps, nil
}

// OAuth2Info struct
type OAuth2Info struct {
	ID      uint           `gorm:"primaryKey"`
	Client  string         `gorm:"not null"` // Client is equal to the idp name
	Info    datatypes.JSON `gorm:"default:'[]'"`
	OauthID string         `gorm:"not null"`
	UserID  string         `gorm:"not null"`
}

func (s *Store) UpdateLarkUserInfo(info OAuth2Info) error {
	return s.db.Table("oauth2_info").
		Where("user_id = ?", info.UserID).
		Where("client = ?", info.Client).
		Update("oauth_user_id = ?", info.OauthID).
		Update("info = ?", info.Info).Error
}

// OauthInfoByUID find user by specific client id in oauth2_info table
//
// return (nil, nil) if user not found
func (s *Store) OauthInfoByUID(clientType, oauthUID string) (*OAuth2Info, error) {
	var client OAuth2Info
	err := s.db.Table("oauth2_info").
		Where("oauth_user_id = ?", oauthUID).
		Where("client = ?", clientType).
		First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to find oauth info by uid")
	}
	return &client, nil
}

// UpsetOauthInfo insert or update oauth2_info table
func (s *Store) UpsetOauthInfo(oauthInfo OAuth2Info) {
	// return Db.Table("oauth2_info").Save(oauthInfo).Error
	stmt := `
	       	INSERT INTO oauth2_info (client, info, oauth_user_id, user_id)
	       	VALUES (?, ?, ?, ?)
	       	ON CONFLICT (client, user_id) DO UPDATE
	       	SET info = EXCLUDED.info, oauth_user_id = EXCLUDED.oauth_user_id, client = EXCLUDED.client
	`

	s.db.Exec(stmt, oauthInfo.Client, oauthInfo.Info, oauthInfo.OauthID, oauthInfo.UserID)
}
