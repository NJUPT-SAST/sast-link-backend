package service

import (
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"gorm.io/datatypes"
)

type OauthService struct {
	*BaseService
}

func NewOauthService(store *BaseService) *OauthService {
	return &OauthService{store}
}

func (s *OauthService) UpsetOauthInfo(username, clientType, oauthID string, OAuth2Info datatypes.JSON) {
	var oauthInfo = store.OAuth2Info{
		Client:  clientType,
		Info:    OAuth2Info,
		OauthID: oauthID,
		UserID:  username,
	}
	s.Store.UpsetOauthInfo(oauthInfo)
}

// Oauth server
func (s *OauthService) OauthUserInfo(userID string) (*store.User, error) {
	return s.Store.UserInfo(userID)
}
