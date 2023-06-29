package service

import (
	"context"
	"fmt"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var serviceLogger = log.Log

func CreateUser(emal string, password string) error{
	err := model.CreateUser(&model.User{
		Email:    &emal,
		Password: &password,
	})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func VerifyAccount(username string, flag string) (string, error) {
	// 0 is register
	// 1 is login
	if flag == "0" {
		return VerifyAccountRegister(username)
	} else if flag == "1" {
		return VerifyAccountLogin(username)
	} else {
		return "", result.ParamError
	}
}

// verify ticket at register
func VerifyAccountRegister(username string) (string, error) {
	// check if the user is exist
	exist, err := model.CheckUserByEmail(username)
	if err != nil {
		return "", err
	}
	// user is exist and can't register
	if exist {
		return "", result.UserIsExist
	} else { // user is not exist and can register
		ticket, err := util.GenerateToken(fmt.Sprintf("%s-register", username))
		if err != nil {
			return "", err
		}
		// 5min expire
		model.Rdb.Set(ctx, "REGISTER_TICKET:"+username, ticket, time.Minute*5)
		return ticket, err
	}
}

// verify ticket at login
func VerifyAccountLogin(username string) (string, error) {
	exist, err := model.CheckUserByEmail(username)
	if err != nil {
		return "", err
	}
	// user is exist and can login
	if exist {
		ticket, err := util.GenerateToken(fmt.Sprintf("%s-login", username))
		if err != nil {
			return "", err
		}
		// 5min expire
		model.Rdb.Set(ctx, "LOGIN_TICKET:"+username, ticket, time.Minute*5)
		return ticket, err
	} else { // user is not exist and can't login
		// login can use uid and email
		uidExist, err := model.CheckUserByUid(username)
		if err != nil {
			return "", err
		}
		if uidExist {
			ticket, err := util.GenerateToken(fmt.Sprintf("%s-login", username))
			if err != nil {
				return "", err
			}
			// 5min expire
			model.Rdb.Set(ctx, "LOGIN_TICKET:"+username, ticket, time.Minute*5)
			return ticket, err
		} else {
			return "", result.UserNotExist
		}
	}
}

func Login(username string, password string) (bool, error) {
	//check password
	flag, err := model.CheckPassword(username, password)
	if !flag {
		return false, err
	}
	return true, err
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
	ticketKey := fmt.Sprintf("REGISTER_TICKET:%s", username)
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
		return result.TICKET_NOT_CORRECT
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
