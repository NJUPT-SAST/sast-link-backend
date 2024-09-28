package service

import (
	"context"

	"gorm.io/datatypes"

	"github.com/NJUPT-SAST/sast-link-backend/store"
)

type OauthService struct {
	*BaseService
}

func NewOauthService(store *BaseService) *OauthService {
	return &OauthService{store}
}

func (s *OauthService) UpsetOauthInfo(username, clientType, oauthID string, oauth2Info datatypes.JSON) {
	var oauthInfo = store.OAuth2Info{
		Client:  clientType,
		Info:    oauth2Info,
		OauthID: oauthID,
		UserID:  username,
	}
	s.Store.UpsetOauthInfo(oauthInfo)
}

// Oauth server.
func (s *OauthService) OauthUserInfo(ctx context.Context, userID string) (*store.User, error) {
	return s.Store.UserInfo(ctx, userID)
}
