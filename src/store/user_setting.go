package store

import (
	"context"
	"encoding/json"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/pkg/errors"
)

type UserSettingKey int32

const (
	USER_SETTING_KEY_UNSPECIFIED UserSettingKey = 0
	// Access tokens for the user.
	USER_SETTING_ACCESS_TOKENS UserSettingKey = 1
)

// Enum value maps for UserSettingKey.
var (
	UserSettingKey_name = map[int32]string{
		0: "USER_SETTING_KEY_UNSPECIFIED",
		1: "USER_SETTING_ACCESS_TOKENS",
	}
	UserSettingKey_value = map[string]int32{
		"USER_SETTING_KEY_UNSPECIFIED": 0,
		"USER_SETTING_ACCESS_TOKENS":   1,
	}
)

// UserSetting represents a user setting.
//
// The value is a JSON string.
type UserSetting struct {
	UserID string         `json:"user_id,omitempty"`
	Key    UserSettingKey `json:"key,omitempty"`
	Value  string         `json:"value,omitempty"`
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

// AccessTokensUserSetting_AccessToken represents an access token for the user.
type AccessTokensUserSetting_AccessToken struct {
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

func (a *AccessTokensUserSetting_AccessToken) String() string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

// AccessTokensUserSetting represents the access tokens for the user.
type AccessTokensUserSetting struct {
	AccessTokens []*AccessTokensUserSetting_AccessToken `json:"access_tokens,omitempty"`
}

func (a *AccessTokensUserSetting) String() string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

func (e UserSettingKey) String() string {
	switch e {
	case USER_SETTING_ACCESS_TOKENS:
		return "USER_SETTING_ACCESS_TOKENS"
	default:
		return "USER_SETTING_KEY_UNSPECIFIED"
	}
}

func (x *UserSetting) GetAccessTokens() *AccessTokensUserSetting {
	var accessTokens AccessTokensUserSetting
	if x != nil {
		err := json.Unmarshal([]byte(x.Value), &accessTokens)
		if err != nil {
			return nil
		}
		return &accessTokens
	}
	return nil
}

func (x *AccessTokensUserSetting) GetAccessTokens() []*AccessTokensUserSetting_AccessToken {
	if x != nil {
		return x.AccessTokens
	}
	return nil
}

func (s *Store) ListUserSettings(ctx context.Context, find *FindUserSetting) ([]*UserSetting, error) {
	var userSettings []*UserSetting
	err := s.db.WithContext(ctx).Table("user_setting").Where("user_id = ?", find.UserID).Find(&userSettings).Error
	if err != nil {
		return nil, err
	}

	for _, userSetting := range userSettings {
		s.Set(ctx, userSetting.UserID, userSetting, 0)
	}

	return userSettings, nil
}

func (s *Store) GetUserSetting(ctx context.Context, find *FindUserSetting) (*UserSetting, error) {
	// Get user setting from cache
	userSettingStr, err := s.Get(ctx, find.UserID)
	if err != nil {
		log.Errorf("Failed to get user setting from cache: %v", err)
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
	s.Set(ctx, find.UserID, userSetting, 0)
	return userSetting, nil
}

func (s *Store) GetUserAccessTokens(ctx context.Context, userID string) ([]*AccessTokensUserSetting_AccessToken, error) {
	userSetting, err := s.GetUserSetting(ctx, &FindUserSetting{
		UserID: userID,
		Key:    USER_SETTING_ACCESS_TOKENS,
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
		Key:    USER_SETTING_ACCESS_TOKENS,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get user setting")
	}

	userAccessToken := &AccessTokensUserSetting_AccessToken{
		AccessToken: accessToken,
		Description: description,
		CreatedTs:   0,
		LastUsedTs:  0,
	}

	if userSetting == nil {
		userSetting = &UserSetting{
			UserID: userID,
			Key:    USER_SETTING_ACCESS_TOKENS,
			Value:  "",
		}
	}

	accessTokens := userSetting.GetAccessTokens()
	if accessTokens == nil {
		accessTokens = &AccessTokensUserSetting{}
	}

	accessTokens.AccessTokens = append(accessTokens.AccessTokens, userAccessToken)
	log.Debugf("User [%s] has %d access tokens", userID, len(accessTokens.AccessTokens))
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
