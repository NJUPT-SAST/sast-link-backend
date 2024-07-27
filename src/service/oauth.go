package service

import (
	"encoding/json"
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/model"
)

// Oauth Github
func GetUserByGithubId(githubId string) (*model.OAuth2Info, error) {
	return model.OauthInfoByUID(model.GITHUB_CLIENT_TYPE, githubId)
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

func UpsetOauthInfo(username, clientType, oauthID, OAuth2Info string) {
	var oauthInfo = model.OAuth2Info{
		Client:  clientType,
		Info:    json.RawMessage(OAuth2Info),
		OauthID: oauthID,
		UserID:  username,
	}
	model.UpsetOauthInfo(oauthInfo)
}

// Oauth Lark
func OauthInfoByLarkID(unionID string) (*model.OAuth2Info, error) {
	return model.OauthInfoByUID(model.LARK_CLIENT_TYPE, unionID)
}

func UserByLarkID(username, unionID string) (*model.User, error) {
	// FIXME: replace union_id with "real" field name in db
	user, err := model.UserByField("lark_id", unionID)
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

// Oauth server
func OauthUserInfo(userID string) (*model.User, error) {
	return model.UserInfo(userID)
}
