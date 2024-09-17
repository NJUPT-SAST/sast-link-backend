package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"
)

type IdentityProviderType string

const (
	// OAuth2 represents the OAuth2 identity provider type.
	IDPTypeOAuth2 IdentityProviderType = "OAuth2"

	IDPTypeUnknown IdentityProviderType = "Unknown"

	OAUTH2_GITHUB = "github"
	OAUTH2_LARK   = "lark"
)

// IdentityProviderUserInfo represents the identity provider user info.
//
// From the identity provider, we can get the user info, such as the identifier, display name, email, etc.
type IdentityProviderUserInfo struct {
	Identifier   string
	IdentifierID string
	DisplayName  string
	Email        string
	Fields       map[string]interface{} // All fields from the identity provider user info.
}

// IdentityProviderSetting represents the identity provider setting.
//
// It container the identity provider setting information, such as the identity provider type, client id, client secret, etc.
type IdentityProviderSetting struct {
	Name string               `json:"name"`
	Type IdentityProviderType `json:"type"` // e.g. OAuth2, OIDC, etc. (Now only support OAuth2)

	// Types that are valid to be assigned to Config:
	//	*OAuth2Setting
	Config isidentityProviderConfig `json:"config"`
}

// UnmarshalJSON implements the custom unmarshalling for IdentityProviderSetting.
func (s *IdentityProviderSetting) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Unmarshal common fields
	if name, ok := raw["name"].(string); ok {
		s.Name = name
	}
	if typ, ok := raw["type"].(string); ok {
		s.Type = IdentityProviderType(typ)
	}

	switch s.Type {
	case IDPTypeOAuth2:
		var config OAuth2Setting
		rawConfig, err := json.Marshal(raw["config"])
		if err != nil {
			return err
		}
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return err
		}
		s.Config = &config
	default:
		return fmt.Errorf("identity provider type %s is not supported", s.Type)
	}

	return nil
}

// func (s *IdentityProviderSetting) MarshalJSON() ([]byte, error) {
// 	switch s.Type {
// 	case IDPTypeOAuth2:
// 		return json.Marshal(struct {
// 			Name   string         `json:"name"`
// 			Type   string         `json:"type"`
// 			Config *OAuth2Setting `json:"config"`
// 		}{
// 			Name:   s.Name,
// 			Type:   string(s.Type),
// 			Config: s.Config.(*OAuth2Setting),
// 		})
// 	default:
// 		return nil, fmt.Errorf("identity provider type %s is not supported", s.Type)
// 	}
// }

// NewIdentityProvider initializes a new OAuth2 Identity Provider with the given configuration.
func NewIdentityProvider(name string, idpType IdentityProviderType, config *isidentityProviderConfig) (*IdentityProviderSetting, error) {
	if idpType != IDPTypeOAuth2 {
		return nil, errors.Errorf("identity provider type %s is not supported", idpType)
	}

	oauth2Config, ok := (*config).(*OAuth2Setting)
	if !ok {
		return nil, errors.Errorf("invalid identity provider config type %T", *config)
	}
	for v, field := range map[string]string{
		oauth2Config.ClientID:                   "clientId",
		oauth2Config.ClientSecret:               "clientSecret",
		oauth2Config.TokenUrl:                   "tokenUrl",
		oauth2Config.UserInfoUrl:                "userInfoUrl",
		oauth2Config.FieldMapping["identifier"]: "fieldMapping.identifier",
	} {
		if v == "" {
			return nil, errors.Errorf(`the field "%s" is empty but required`, field)
		}
	}

	return &IdentityProviderSetting{
		Name:   name,
		Type:   idpType,
		Config: oauth2Config,
	}, nil
}

func (s *IdentityProviderSetting) GetOauth2Setting() *OAuth2Setting {
	if provider, ok := s.Config.(*OAuth2Setting); ok {
		return provider
	}
	return nil
}

// isidentityProviderConfig is an interface that all identity provider config types must implement.
type isidentityProviderConfig interface {
	isIdentityProviderConfig()
}

// Oauth2IdentityProvider represents the OAuth2 identity provider.
type Oauth2IdentityProvider struct {
	config *OAuth2Setting

	// ExchangeTokenFunc is the function to exchange the OAuth2 token.
	ExchangeTokenFunc func(ctx context.Context, oauth2Setting *OAuth2Setting, redirectURL, code string) (string, error)
	// UserInfoFunc is the function to get the user info from the identity provider.
	UserInfoFunc func(ctx context.Context, oauth2Setting *OAuth2Setting, token string) (*IdentityProviderUserInfo, error)
}

func NewOauth2IdentityProvider(idpSetting *IdentityProviderSetting) (*Oauth2IdentityProvider, error) {
	config := idpSetting.GetOauth2Setting()
	for v, field := range map[string]string{
		config.ClientID:                   "clientId",
		config.ClientSecret:               "clientSecret",
		config.TokenUrl:                   "tokenUrl",
		config.UserInfoUrl:                "userInfoUrl",
		config.FieldMapping["identifier"]: "fieldMapping.identifier",
	} {
		if v == "" {
			return nil, errors.Errorf(`the field "%s" is empty but required`, field)
		}
	}

	var exchangeTokenFunc func(ctx context.Context, oauth2Setting *OAuth2Setting, redirectURL, code string) (string, error)
	var userInfoFunc func(ctx context.Context, oauth2Setting *OAuth2Setting, token string) (*IdentityProviderUserInfo, error)
	switch {
	case idpSetting.Name == OAUTH2_LARK:
		exchangeTokenFunc = LarkExchangeToken
		userInfoFunc = LarkUserInfo
	default:
		exchangeTokenFunc = ExchangeToken
		userInfoFunc = UserInfo
	}

	return &Oauth2IdentityProvider{
		config:            config,
		ExchangeTokenFunc: exchangeTokenFunc,
		UserInfoFunc:      userInfoFunc,
	}, nil
}

func (o *Oauth2IdentityProvider) GetConfig() *OAuth2Setting {
	return o.config
}

// isIdentityProviderConfig is an interface that all identity provider config types must implement.
func (o *Oauth2IdentityProvider) isIdentityProviderConfig() {}

// ExchangeToken returns the exchanged OAuth2 token using the given authorization code.
func (o *Oauth2IdentityProvider) ExchangeToken(ctx context.Context, oauth2Setting *OAuth2Setting, redirectURL, code string) (string, error) {
	return o.ExchangeTokenFunc(ctx, oauth2Setting, redirectURL, code)
}

// UserInfo returns the user info from the identity provider.
func (o *Oauth2IdentityProvider) UserInfo(ctx context.Context, oauth2Setting *OAuth2Setting, accessToken string) (*IdentityProviderUserInfo, error) {
	return o.UserInfoFunc(ctx, oauth2Setting, accessToken)
}

// OAuth2Setting represents the OAuth2 setting.
//
// It container the OAuth2 setting information, such as the client id, client secret, auth url, etc.
type OAuth2Setting struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthUrl      string   `json:"auth_url"`
	TokenUrl     string   `json:"token_url"`
	UserInfoUrl  string   `json:"user_info_url"`
	Scopes       []string `json:"scopes"`
	// FieldMapping is the mapping between the identity provider user info and the system user info.
	// eg. {"id": "identifier", "name": "display_name", "email": "email"}, the key is the field name in the identity provider user info,
	// the value is the field name in the system user info.
	FieldMapping map[string]string `json:"field_mapping"`
}

func (c *OAuth2Setting) isIdentityProviderConfig() {}

// ToJSON converts the OAuth2Setting to JSON string.
func (s *OAuth2Setting) ToJSON() (string, error) {
	jsonStr, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonStr), nil
}

// FromJSON converts the JSON string to OAuth2Setting.
func (s *OAuth2Setting) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), s)
}

// ExchangeToken returns the exchanged OAuth2 token using the given authorization code.
func ExchangeToken(ctx context.Context, oauth2Setting *OAuth2Setting, redirectURL, code string) (string, error) {
	log.Debugf("ExchangeToken::[redirectURL: %s] [code: %s]", redirectURL, code)
	conf := &oauth2.Config{
		ClientID:     oauth2Setting.ClientID,
		ClientSecret: oauth2Setting.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       oauth2Setting.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   oauth2Setting.AuthUrl,
			TokenURL:  oauth2Setting.TokenUrl,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	token, err := conf.Exchange(ctx, code)
	log.Debug("ExchangeToken::", err)
	if err != nil {
		return "", errors.Wrap(err, "failed to exchange access token")
	}

	accessToken, ok := token.Extra("access_token").(string)
	log.Debugf("ExchangeToken::[accessToken: %s]", accessToken)
	if !ok {
		return "", errors.New(`missing "access_token" from authorization response`)
	}

	return accessToken, nil
}

func UserInfo(ctx context.Context, oauth2Setting *OAuth2Setting, token string) (*IdentityProviderUserInfo, error) {
	client := &http.Client{}
	log.Debug("UserInfo::userInfoUrl: ", oauth2Setting.UserInfoUrl)
	req, err := http.NewRequest(http.MethodGet, oauth2Setting.UserInfoUrl, nil)
	if err != nil {
		log.Debug("UserInfo::", err)
		return nil, errors.Wrap(err, "failed to new http request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		log.Debug("UserInfo::", err)
		return nil, errors.Wrap(err, "failed to get user information")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debug("UserInfo::", err)
		return nil, errors.Wrap(err, "failed to read response body")
	}
	defer resp.Body.Close()

	var claims map[string]any
	err = json.Unmarshal(body, &claims)
	log.Debug("UserInfo::body", string(body))
	if err != nil {
		log.Debug("UserInfo::", err)
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	userInfo := mapClaimsToUserInfo(claims, oauth2Setting.FieldMapping)
	return userInfo, nil
}

func LarkExchangeToken(ctx context.Context, oauth2Setting *OAuth2Setting, redirectURL, code string) (string, error) {
	appAccessToken, err := larkAppAccessToken(oauth2Setting)
	if err != nil || appAccessToken == "" {
		return "", err
	}

	data := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}

	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", appAccessToken),
		"Content-Type":  "application/json; charset=utf-8",
	}
	res, err := util.PostWithHeader(oauth2Setting.TokenUrl, header, data)
	if err != nil {
		log.Errorf("util.PostWithHeader ::: %s", err.Error())
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Errorf("io.ReadAll ::: %s", err.Error())
		return "", errors.Wrap(err, "failed to read response body")
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Errorf("LarkExchangeToken ::: gjson.Get ::: response code: %d\n", resCode)
		return "", errors.New("Failed to exchange token")
	}

	userAccessToken := gjson.Get(string(body), "data.access_token").String()
	if userAccessToken == "" {
		return "", errors.New(`Missing "access_token" from authorization response`)
	}

	return userAccessToken, nil
}

func LarkUserInfo(ctx context.Context, oauth2Setting *OAuth2Setting, accessToken string) (*IdentityProviderUserInfo, error) {
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	res, err := util.GetWithHeader(oauth2Setting.UserInfoUrl, header)
	if err != nil {
		log.Errorf("util.GetWithHeader ::: %s", err.Error())
		return nil, errors.Wrap(err, "failed to get user information")
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Errorf("io.ReadAll ::: %s", err.Error())
		return nil, errors.Wrap(err, "failed to read response body")
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Errorf("LarkUserInfo ::: gjson.Get ::: response code: %d\n", resCode)
		return nil, errors.New("Failed to get user information")
	}

	var claims map[string]any
	err = json.Unmarshal(body, &claims)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	userInfo := mapClaimsToUserInfo(claims, oauth2Setting.FieldMapping)
	return userInfo, nil
}

// larkAppAccessToken returns the app_access_token from lark
func larkAppAccessToken(oauth2Setting *OAuth2Setting) (string, error) {
	appId := oauth2Setting.ClientID
	appSecret := oauth2Setting.ClientSecret

	params := url.Values{}
	params.Add("app_id", appId)
	params.Add("app_secret", appSecret)

	appAccessTokenURL := "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"

	res, error := http.PostForm(appAccessTokenURL, params)
	if error != nil {
		log.Errorf("http.PostForm ::: %s", error.Error())
		return "", error
	}
	// log.LogRes(res)

	body, error := io.ReadAll(res.Body)
	defer res.Body.Close()
	if error != nil {
		log.Errorf("io.ReadAll ::: %s", error.Error())
		return "", error
	}

	if code := gjson.Get(string(body), "code").Int(); code != 0 {
		log.Errorf("gjson.Get ::: code: %d", code)
		return "", errors.New("Failed to get app_access_token")
	}

	acceToken := gjson.Get(string(body), "app_access_token").String()
	expire := gjson.Get(string(body), "expire").Int()
	log.Infof("larkAppAccessToken ::: expire: %d", expire)

	// TODO: store app_access_token in redis/postgresql
	// model.Rdb.Set(model.RedisCtx, "lark_app_access_token", acceToken, time.Duration(expire))

	return acceToken, nil
}

// mapClaimsToUserInfo maps the claims to the identity provider user info.
func mapClaimsToUserInfo(claims map[string]any, fieldMapping map[string]string) *IdentityProviderUserInfo {
	userInfo := &IdentityProviderUserInfo{
		Fields:      claims,
		Identifier:  extractStringField(claims, fieldMapping, "identifier"),
		DisplayName: extractStringField(claims, fieldMapping, "display_name"),
		Email:       extractStringField(claims, fieldMapping, "email"),
	}

	if userInfo.Identifier == "" {
		return nil
	}

	if userInfo.DisplayName == "" {
		userInfo.DisplayName = userInfo.Identifier
	}

	return userInfo
}

func extractStringField(claims map[string]any, fieldMapping map[string]string, fieldKey string) string {
	if fieldName, exists := fieldMapping[fieldKey]; exists && fieldName != "" {
		if value, ok := claims[fieldName].(string); ok {
			return value
		}
	}
	return ""
}
