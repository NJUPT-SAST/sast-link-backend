package store

import (
	"encoding/json"
	"errors"
)

type UserSettingKey int32

const (
	UserSettingKey_USER_SETTING_KEY_UNSPECIFIED UserSettingKey = 0
	// Access tokens for the user.
	UserSettingKey_USER_SETTING_ACCESS_TOKENS UserSettingKey = 1
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

type UserSetting struct {
	UserID string
	Key    UserSettingKey
	Value  string
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
	case UserSettingKey_USER_SETTING_ACCESS_TOKENS:
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

func (s *Store) ListUserSettings(find *FindUserSetting) ([]*UserSetting, error) {
	var userSettings []*UserSetting
	err := s.db.Table("user_setting").Where("user_id = ?", find.UserID).Find(&userSettings).Error
	if err != nil {
		return nil, err
	}
	return userSettings, nil
}

func (s *Store) GetUserSetting(find *FindUserSetting) (*UserSetting, error) {
	list, err := s.ListUserSettings(find)
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
	//TODO: Cache the user setting.
	return userSetting, nil
}

func (s *Store) GetUserAccessTokens(userID string) ([]*AccessTokensUserSetting_AccessToken, error) {
    userSetting, err := s.GetUserSetting(&FindUserSetting{
        UserID: userID,
        Key: UserSettingKey_USER_SETTING_ACCESS_TOKENS,
    })
    if err != nil {
        return nil, err
    }
    if userSetting == nil {
        return nil, nil
    }
    return userSetting.GetAccessTokens().GetAccessTokens(), nil
}
