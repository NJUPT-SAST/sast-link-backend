package service

import (
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/model"
)

func OauthUserInfo(userID string) (*model.User, error) {
	return model.UserInfo(userID)
}

func GetUserInfoFromGithub(username, githubId string) (*model.User, error) {
	user, err := model.FindUserByGithubId(githubId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Uid != &username {
		return nil, fmt.Errorf("user not match")
	}
	return nil, nil
}
