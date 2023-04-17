package service

import (
	"github.com/NJUPT-SAST/sast-link-backend/model"
)

func CreateUser(emal string, password string) {
	model.CreateUser(&model.User{
		Email:    &emal,
		Password: &password,
	})
}
