package service

import (
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/model"
)


// Oauth Github
func GetUserByGithubId(githubId string) (*model.User, error) {
	return model.UserByField("github_id", githubId)
}

func GetUserInfoFromGithub(username, githubId string) (*model.User, error) {
	user, err := model.UserByField("github_id", githubId)
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

// Oauth Lark
func UserByLarkUnionID(unionID string) (*model.User, error) {
	return model.UserByField("lark_id", unionID)
}

func UpdateLarkUserInfo(username, clientType, oauthID, larkUserInfo string) error {
	return model.UpdateLarkUserInfo(username, clientType, oauthID, larkUserInfo)
}

// Oauth server
func OauthUserInfo(userID string) (*model.User, error) {
	return model.UserInfo(userID)
}
