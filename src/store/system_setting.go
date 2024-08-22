package store

import (
	"context"
	"encoding/json"


	"github.com/NJUPT-SAST/sast-link-backend/plugin/idp/oauth2"
)

// SystemSettingType represents the system setting type.
type SystemSettingType string

const (
	// WebsiteSettingType represents the website setting type.
	WebsiteSettingType SystemSettingType = "website"
	// EmailSettingType represents the email setting type.
	EmailSettingType SystemSettingType = "email"
	// StorageSettingType represents the storage setting type.
	StorageSettingType SystemSettingType = "storage"
	// IdpSettingType represents the identity provider setting type.
	IdpSettingType SystemSettingType = "idp"
)

// String converts the SystemSettingType to string.
func (t SystemSettingType) String() string {
	return string(t)
}

func TypeFromString(t string) SystemSettingType {
	switch t {
	case "website":
		return WebsiteSettingType
	case "email":
		return EmailSettingType
	case "storage":
		return StorageSettingType
	case "idp":
		return IdpSettingType
	}
	return ""
}

// SystemSetting represents the system setting.
//
// It container the system setting information, such as the email sender and secret key, etc.
type SystemSetting struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// GetSettings return the system setting by type.
func (s *SystemSetting) GetSettings() (interface{}, error) {
	switch SystemSettingType(s.Name) {
	case WebsiteSettingType:
		var setting WebsiteSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	case EmailSettingType:
		var setting EmailSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	case StorageSettingType:
		var setting StorageSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	case IdpSettingType:
		var setting oauth2.IdentityProviderSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	}
	return nil, nil
}

// GetWebsiteSetting get the website setting. If the system setting is not a website setting, return nil.
func (s *SystemSetting) GetWebsiteSetting() (*WebsiteSetting, error) {
	if s.Name != WebsiteSettingType.String() {
		return nil, nil
	}
	var setting WebsiteSetting
	if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetEmailSetting get the email setting. If the system setting is not an email setting, return nil.
func (s *SystemSetting) GetEmailSetting() (*EmailSetting, error) {
	if s.Name != EmailSettingType.String() {
		return nil, nil
	}
	var setting EmailSetting
	if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetStorageSetting get the storage setting. If the system setting is not a storage setting, return nil.
func (s *SystemSetting) GetStorageSetting() (*StorageSetting, error) {
	if s.Name != StorageSettingType.String() {
		return nil, nil
	}
	var setting StorageSetting
	if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetIdpSetting get the identity provider setting. If the system setting is not an identity provider setting, return nil.
func (s *SystemSetting) GetIdpSetting() (*oauth2.IdentityProviderSetting, error) {
	if s.Name != IdpSettingType.String() {
		return nil, nil
	}
	var setting oauth2.IdentityProviderSetting
	if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

// WebsiteSetting represents the website setting.
//
// It container the website setting information, such as the website name, description, allowRegister, etc.
type WebsiteSetting struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AllowRegister bool   `json:"allowRegister"`
	// URL of the image to be displayed when an error occurs
	ResponseErrorURLImage string `json:"response_error_url_image"`
	AvatarErrorURLImage   string `json:"avatar_error_url_image"`
	// URL of the frontend
	FrontendURL string `json:"frontend_url"`
}

// EmailSetting represents the email setting.
//
// It container the email setting information, such as the email sender and secret key, etc.
type EmailSetting struct {
	Sender string
	Secret string
}

// StorageSetting represents the storage setting.
//
// It container the storage setting information, such as the storage type, access key, secret key, etc.
type StorageSetting struct {
	Type      string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Endpoint  string
}

// UpsetSystemSetting update or insert the system setting.
func (s *Store) UpsetSystemSetting(ctx context.Context, setting *SystemSetting) error {
	s.db.Table("system_setting").Save(setting)

	// Update the system setting in the cache
	return s.Set(ctx, setting.Name, setting, 0)
}

// ListSystemSetting list the system setting.
func (s *Store) ListSystemSetting(ctx context.Context) (map[SystemSettingType]any, error) {
	settings := make(map[SystemSettingType]any)
	s.db.Table("system_setting").Find(&settings)

	for k, setting := range settings {
		// Store the system setting in the cache
		s.Set(ctx, k.String(), setting, 0)
	}

	return settings, nil
}

// GetSystemSetting get the system setting by name.
func (s *Store) GetSystemSetting(ctx context.Context, settingType SystemSettingType) (*SystemSetting, error) {
	// Get the system setting from the cache
	settingStr, err := s.Get(ctx, settingType.String())
	if err != nil {
		return nil, err
	}

	// If the system setting exists in the cache, return it
	if settingStr != "" {
		var setting SystemSetting
		if err := json.Unmarshal([]byte(settingStr), &setting); err != nil {
			return nil, err
		}
		return &setting, nil
	}

	// If the system setting does not exist in the cache, get it from the database
	var setting SystemSetting
	s.db.Table("system_setting").Where("name = ?", settingType.String()).First(&setting)
	if err := s.Set(ctx, settingType.String(), &setting, 0); err != nil {
		return nil, err
	}

	return &setting, nil
}
