package service

import (
	"context"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/redis/go-redis/v9"
)

var serviceLogger = log.Log
var ctx = context.Background()

func CreateUser(emal string, password string) {
	model.CreateUser(&model.User{
		Email:    &emal,
		Password: &password,
	})
}

func Login(username string, password string) (bool, error) {
	//check password
	flag, err := model.CheckPassword(username, password)
	if !flag {
		return false, err
	}
	return true, err
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
	ticketKey := "TICKET:" + username
	ctx := context.Background()
	val, err := model.Rdb.Get(ctx, ticketKey).Result()
	if err != nil {
		// key does not exists
		if err == redis.Nil {
			return result.CHECK_TICKET_NOTFOUND.Wrap(err)
		}
		return err
	}
	// ticket is not correct
	if val != ticket {
		return result.TICKET_NOT_CORRECT.Wrap(err)
	}
	code := model.GenerateVerifyCode(username)
	codeKey := "VERIFY_CODE:" + username
	// 3min expire
	model.Rdb.Set(ctx, codeKey, code, time.Minute*3)
	serviceLogger.Infof("Send Email to [%s] with code [%s]\n", username, code)
	content := model.InsertCode(code)
	emailErr := model.SendEmail(username, content)
	if emailErr != nil {
		return emailErr
	}
	return nil
}

func CheckVerifyCode(username string, code string) error {
	key := "VERIFY_CODE:" + username
	val, err := model.Rdb.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return result.VerifyCodeError
		}
		return err
	}
	if val != code {
		return result.VerifyCodeError
	}
	return nil
}
