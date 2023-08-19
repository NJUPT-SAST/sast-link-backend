package service

import "github.com/NJUPT-SAST/sast-link-backend/model"

func OauthUserInfo(userID string) (*model.User, error) {
	return model.UserInfo(userID)
}
