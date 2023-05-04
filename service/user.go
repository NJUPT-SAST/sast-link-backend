package service

import (
	"context"
	"fmt"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/redis/go-redis/v9"
)

var serviceLogger = log.Log

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
	username, claimsError := jwtClaims.GetSubject()
	if claimsError != nil {
		return nil, claimsError
	}
	return model.UserInfo(username)
}

func SendEmail(username string, ticket string) error {
	key := "TICKET:" + username
	val, err := model.Rdb.Get(context.Background(), key).Result()
	if err != nil {
		// key does not exists
		if err == redis.Nil {
			return fmt.Errorf(result.GetMsg(result.ERROR_CHECK_TICKET_NOTFOUND), err)
		}
		return err
	}
	// ticket is not correct
	if val != ticket {
		return fmt.Errorf(result.GetMsg(result.ERROR_CHECK_TICKET_NOTFOUND), err)
	}
	code := model.GenerateVerifyCode(username)
	emailErr := model.SendEmail(username, code)
	if emailErr != nil {
		return emailErr
	}
	return nil
}
