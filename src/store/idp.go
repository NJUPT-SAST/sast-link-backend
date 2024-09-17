package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/plugin/idp/oauth2"
)

func (s *Store) CreateIdentityProvider(ctx context.Context, idp *oauth2.IdentityProviderSetting) error {
	// Check if the identity provider already exists
	if _, err := s.GetIdentityProviderByName(ctx, idp.Name); err == nil {
		return errors.Errorf("identity provider %s already exists", idp.Name)
	}
	if err := s.db.Table("idp").Create(idp).Error; err != nil {
		return err
	}

	return s.Set(ctx, idp.Name, idp, 0)
}

func (s *Store) GetIdentityProviderByName(ctx context.Context, name string) (*oauth2.IdentityProviderSetting, error) {
	if !strings.HasPrefix(name, config.IdpSettingType.String()) {
		name = fmt.Sprintf("%s-%s", config.IdpSettingType.String(), name)
	}
	var idpSetting *SystemSetting
	idpCache, err := s.Get(ctx, name)
	if err == nil && idpCache != "" {
		if err := json.Unmarshal([]byte(idpCache), &idpSetting); err != nil {
			return nil, err
		}
		if idpSetting.GetIdpSetting() == nil {
			return idpSetting.GetIdpSetting(), errors.Errorf("idp setting with name: [%s] not exits", name)
		}
		return idpSetting.GetIdpSetting(), nil
	}

	idpSetting, err = s.GetSystemSetting(ctx, name)
	if err != nil {
		return nil, err
	}
	return idpSetting.GetIdpSetting(), nil
}

// ListIdentityProviders returns all identity providers
func (s *Store) ListIdentityProviders(ctx context.Context) ([]oauth2.IdentityProviderSetting, error) {
	var idps []oauth2.IdentityProviderSetting
	// FIX: cache cna't update when add new idp
	// idpCache, err := s.Get(ctx, config.IdpSettingType.String())
	// if err == nil && idpCache != "" {
	// 	json.Unmarshal([]byte(idpCache), &idps)
	// 	return idps, nil
	// }

	systemSetting, err := s.ListSystemSetting(ctx)
	if err != nil {
		return idps, err
	}

	for _, setting := range systemSetting {
		if setting.Type == config.IdpSettingType.String() {
			idp := setting.GetIdpSetting()
			if idp != nil {
				s.Set(ctx, idp.Name, idp, 0)
				idps = append(idps, *idp)
			}
		}
	}

	s.Set(ctx, config.IdpSettingType.String(), idps, 0)
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
