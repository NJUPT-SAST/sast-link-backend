package store

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type UserSettingKey int32

const (
	UserSettingKeyUnspecified = iota
	// Access tokens for the user.
	UserSettingAccessTokens
)

// Enum value maps for UserSettingKey.
var (
	UserSettingKeyName = map[int32]string{
		UserSettingKeyUnspecified: "USER_SETTING_KEY_UNSPECIFIED",
		UserSettingAccessTokens:   "USER_SETTING_ACCESS_TOKENS",
	}
	UserSettingKeyValue = map[string]int32{
		"USER_SETTING_KEY_UNSPECIFIED": UserSettingKeyUnspecified,
		"USER_SETTING_ACCESS_TOKENS":   UserSettingAccessTokens,
	}
)

func (e UserSettingKey) String() string {
	switch e {
	case UserSettingAccessTokens:
		return "USER_SETTING_ACCESS_TOKENS"
	default:
		return "USER_SETTING_KEY_UNSPECIFIED"
	}
}

// UserSetting represents a user setting.
//
// The value is a JSON string.
type UserSetting struct {
	UserID string         `json:"user_id,omitempty"`
	Key    UserSettingKey `json:"key,omitempty"`
	Value  string         `json:"value,omitempty"`
}

// GetAccessTokens returns the access tokens for the user setting.
//
// AccessToeknsUserSetting is a wrapper for the access tokens. It container a slice of UserSetting_AccessToken.
func (s *UserSetting) GetAccessTokens() *AccessTokensUserSetting {
	var accessTokens AccessTokensUserSetting
	if s != nil {
		err := json.Unmarshal([]byte(s.Value), &accessTokens)
		if err != nil {
			return nil
		}
		return &accessTokens
	}
	return nil
}

func (s *UserSetting) String() string {
	j, _ := json.Marshal(s)
	return string(j)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s UserSetting) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *UserSetting) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type FindUserSetting struct {
	UserID string
	Key    UserSettingKey
}

// UserSettingAccessToken represents an access token for the user.
type UserSettingAccessToken struct {
	// The access token is a JWT token.
	// Including expiration time, issuer, etc.
	AccessToken string `json:"access_token,omitempty"`
	// A description for the access token.
	Description string `json:"description,omitempty"`
	// The time when the access token was created.
	CreatedTs int64 `json:"created_ts,omitempty"`
	// The time when the access token was last used.
	LastUsedTs int64 `json:"last_used_ts,omitempty"`
}

func (a *UserSettingAccessToken) String() string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

// AccessTokensUserSetting represents the access tokens for the user.
//
// AccessTokensUserSetting is a wrapper for the access tokens.
type AccessTokensUserSetting struct {
	AccessTokens []*UserSettingAccessToken `json:"access_tokens,omitempty"`
}

func (s *AccessTokensUserSetting) String() string {
	if s == nil {
		return ""
	}
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *AccessTokensUserSetting) GetAccessTokens() []*UserSettingAccessToken {
	if s != nil {
		return s.AccessTokens
	}
	return nil
}

// ListUserSettings returns the user settings for the user.
func (s *Store) ListUserSettings(ctx context.Context, find *FindUserSetting) ([]*UserSetting, error) {
	var userSettings []*UserSetting
	err := s.db.WithContext(ctx).Table("user_setting").Where("user_id = ?", find.UserID).Find(&userSettings).Error
	if err != nil {
		return nil, err
	}

	for _, userSetting := range userSettings {
		_ = s.Set(ctx, userSetting.UserID, userSetting, 0)
	}

	return userSettings, nil
}

// GetUserSetting returns the user setting for the user.
func (s *Store) GetUserSetting(ctx context.Context, find *FindUserSetting) (*UserSetting, error) {
	// Get user setting from cache
	userSettingStr, err := s.Get(ctx, find.UserID)
	if err != nil {
		s.log.Error("Failed to get user setting from cache", zap.Error(err))
		return nil, err
	}

	if userSettingStr != "" {
		var userSetting UserSetting
		if err := json.Unmarshal([]byte(userSettingStr), &userSetting); err != nil {
			return nil, err
		}
		return &userSetting, nil
	}

	list, err := s.ListUserSettings(ctx, find)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, nil
	}
	if len(list) > 1 {
		return nil, errors.New("more than one user setting found")
	}

	userSetting := list[0]
	_ = s.Set(ctx, find.UserID, userSetting, 0)
	return userSetting, nil
}

// GetUserAccessTokens returns the access tokens for the user.
func (s *Store) GetUserAccessTokens(ctx context.Context, userID string) ([]*UserSettingAccessToken, error) {
	userSetting, err := s.GetUserSetting(ctx, &FindUserSetting{
		UserID: userID,
		Key:    UserSettingAccessTokens,
	})
	if err != nil {
		return nil, err
	}
	if userSetting == nil {
		return nil, nil
	}
	return userSetting.GetAccessTokens().GetAccessTokens(), nil
}

func (s *Store) UpsetAccessTokensUserSetting(ctx context.Context, userID string, accessToken, description string) error {
	userSetting, err := s.GetUserSetting(ctx, &FindUserSetting{
		UserID: userID,
		Key:    UserSettingAccessTokens,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get user setting")
	}

	userAccessToken := &UserSettingAccessToken{
		AccessToken: accessToken,
		Description: description,
		CreatedTs:   0,
		LastUsedTs:  0,
	}

	if userSetting == nil {
		userSetting = &UserSetting{
			UserID: userID,
			Key:    UserSettingAccessTokens,
			Value:  "",
		}
	}

	accessTokens := userSetting.GetAccessTokens()
	if accessTokens == nil {
		accessTokens = &AccessTokensUserSetting{}
	}

	accessTokens.AccessTokens = append(accessTokens.AccessTokens, userAccessToken)
	userSetting.Value = accessTokens.String()

	return s.UpsetUserSetting(ctx, userSetting)
}

func (s *Store) UpsetUserSetting(ctx context.Context, userSetting *UserSetting) error {
	// Perform upsert operation
	err := s.db.Table("user_setting").Where("user_id = ? AND key = ?", userSetting.UserID, userSetting.Key).Assign(userSetting).Updates(userSetting).Error
	if err != nil {
		return err
	}

	// Update the user setting in the cache
	return s.Set(ctx, userSetting.UserID, userSetting, 0)
}
