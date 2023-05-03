package service

import (
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/util"
)

func CreateUser(emal string, password string) {
	model.CreateUser(&model.User{
		Email:    &emal,
		Password: &password,
	})
}

func VerifyAccount(username string) (bool, string, error) {
	return model.VerifyAccount(username)
}

func UserInfo(jwt string) (*model.User, error) {
	jwtClaims, err := util.ParseToken(jwt)
	if err != nil {
		return nil, err
	}
	return model.UserInfo(jwtClaims.Username)
}
