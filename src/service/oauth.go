package service

import (
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/model"
	"gorm.io/datatypes"
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

func UpsetOauthInfo(username, clientType, oauthID string, OAuth2Info datatypes.JSON) {
	var oauthInfo = model.OAuth2Info{
		Client:  clientType,
		Info:    OAuth2Info,
		OauthID: oauthID,
		UserID:  username,
	}
	model.UpsetOauthInfo(oauthInfo)
}

// Oauth Lark
func UserByLarkUnionID(unionID string) (*model.User, error) {
	return model.UserByField("lark_id", unionID)
}

// Oauth server
func OauthUserInfo(userID string) (*model.User, error) {
	return model.UserInfo(userID)
}
