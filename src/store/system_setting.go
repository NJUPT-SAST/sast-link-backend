package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/plugin/idp/oauth2"
	"gorm.io/gorm"
)

// SystemSetting represents the system setting.
//
// It container the system setting information, such as the email sender and secret key, etc.
type SystemSetting struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (s *SystemSetting) String() string {
	j, _ := json.Marshal(s)
	return string(j)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s SystemSetting) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *SystemSetting) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

// GetSettings return the system setting by type.
func (s *SystemSetting) GetSettings() (interface{}, error) {
	switch config.SystemSettingType(s.Name) {
	case config.WebsiteSettingType:
		var setting WebsiteSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	case config.EmailSettingType:
		var setting EmailSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	case config.StorageSettingType:
		var setting StorageSetting
		if err := json.Unmarshal([]byte(s.Value), &setting); err != nil {
			return nil, err
		}
		return setting, nil
	case config.IdpSettingType:
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
	if s.Name != config.WebsiteSettingType.String() {
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
	if s.Name != config.EmailSettingType.String() {
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
	if s.Name != config.StorageSettingType.String() {
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
	if s.Name != config.IdpSettingType.String() {
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
	// URL of the ueer avatar to be displayed when an error occurs
	AvatarErrorURLImage string `json:"avatar_error_url_image"`
	// URL of the frontend
	FrontendURL string `json:"frontend_url"`
	// Wbhook URL
	// URL of the webhook to be called when the image is need to be reviewed
	ImageReviewWebhook string `json:"image_review_webhook"`
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
	AppID     string
}

// InitSystemSetting initialize the system setting.
//
// It will read the system setting from the config file and store it in the database.
// If the system setting already exists in the database, it will be updated.
func (s *Store) InitSystemSetting(ctx context.Context, profile *config.Config) error {
	// Initialize the system setting
	settings := make(map[config.SystemSettingType]SystemSetting)
	settings[config.WebsiteSettingType] = SystemSetting{
		Name:        config.WebsiteSettingType.String(),
		Value:       profile.SystemSettings[config.WebsiteSettingType.String()],
		Description: "Website setting",
	}
	settings[config.EmailSettingType] = SystemSetting{
		Name:        config.EmailSettingType.String(),
		Value:       profile.SystemSettings[config.EmailSettingType.String()],
		Description: "Email setting",
	}
	settings[config.StorageSettingType] = SystemSetting{
		Name:        config.StorageSettingType.String(),
		Value:       profile.SystemSettings[config.StorageSettingType.String()],
		Description: "Storage setting",
	}
	settings[config.IdpSettingType] = SystemSetting{
		Name:        config.IdpSettingType.String(),
		Value:       profile.SystemSettings[config.IdpSettingType.String()],
		Description: "Identity provider setting",
	}

	// Insert the system setting into the database
	for _, setting := range settings {
		if err := s.InsertSystemSetting(ctx, &setting); err != nil {
			return err
		}
	}

	return nil
}

// UpsetSystemSetting update or insert the system setting.
func (s *Store) UpsetSystemSetting(ctx context.Context, setting *SystemSetting) error {
	// Perform upsert operation
	err := s.db.Table("system_setting").Where("name = ?", setting.Name).Assign(setting).FirstOrCreate(&setting).Error
	if err != nil {
		log.Errorf("Failed to upsert system setting: %s", err.Error())
		return err
	}

	// Update the system setting in the cache
	return s.Set(ctx, setting.Name, setting, 0)
}

// InsertSystemSetting will insert the system setting into the database.
//
// If the system setting already exists in the database, it will be ignored.
func (s *Store) InsertSystemSetting(ctx context.Context, setting *SystemSetting) error {
	var existingSetting SystemSetting

	// Check if the record exists
	err := s.db.Table("system_setting").Where("name = ?", setting.Name).First(&existingSetting).Error
	if err == nil {
		// Record exists, log and cache
		log.Infof("Record with name %s already exists, skipping insert.", setting.Name)
		return s.Set(ctx, existingSetting.Name, existingSetting, 0)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// An error occurred other than "record not found"
		log.Errorf("Failed to check if record exists: %s", err.Error())
		return err
	}

	// Record does not exist, proceed with the insert
	if err := s.db.Table("system_setting").Create(setting).Error; err != nil {
		log.Errorf("Failed to insert system setting: %s", err.Error())
		return err
	}

	// Update the system setting in the cache
	return s.Set(ctx, setting.Name, setting, 0)
}

// ListSystemSetting list the system setting.
func (s *Store) ListSystemSetting(ctx context.Context) (map[config.SystemSettingType]any, error) {
	settings := make(map[config.SystemSettingType]any)
	s.db.Table("system_setting").Find(&settings)

	for k, setting := range settings {
		// Store the system setting in the cache
		s.Set(ctx, k.String(), setting, 0)
	}

	return settings, nil
}

// GetSystemSetting get the system setting by name.
func (s *Store) GetSystemSetting(ctx context.Context, settingType config.SystemSettingType) (*SystemSetting, error) {
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
	if err := s.Set(ctx, settingType.String(), setting, 0); err != nil {
		return nil, err
	}

	return &setting, nil
}
