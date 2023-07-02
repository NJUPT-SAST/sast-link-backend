package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var serviceLogger = log.Log

func CreateUser(email string, password string) error {
	split := regexp.MustCompile(`@`)
	uid := split.Split(email, 2)[0]
	err := model.CreateUser(&model.User{
		Email:    &email,
		Password: &password,
		Uid:      &uid,
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
	} else {
		return VerifyAccountLogin(username)
	}
}

// this function is used to verify the user's email is exist or not when register
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
		// generate token and set expire time
		ticket, err := util.GenerateTokenWithExpireTime(fmt.Sprintf("%s-register", username), model.REGISTER_TICKET_EXPIRE_TIME)
		if err != nil {
			return "", err
		}
		// set token to redis
		model.Rdb.Set(ctx, ticket, model.REGISTER_STATUS["VERIFY_ACCOUNT"], model.REGISTER_TICKET_EXPIRE_TIME)
		return ticket, err
	}
}

// this function is used to verify the user's email is exist or not when login
func VerifyAccountLogin(username string) (string, error) {
	exist, err := model.CheckUserByEmail(username)
	if err != nil {
		return "", err
	}
	// user is exist and can login
	if exist {
		ticket, err := util.GenerateTokenWithExpireTime(fmt.Sprintf("%s-login", username), model.LOGIN_TICKET_EXPIRE)
		if err != nil {
			return "", err
		}
		// 5min expire
		model.Rdb.Set(ctx, "LOGIN_TICKET:"+username, ticket, model.LOGIN_TICKET_EXPIRE)
		return ticket, err
	} else { // user is not exist and can't login
		// login can use uid and email
		uidExist, err := model.CheckUserByUid(username)
		if err != nil {
			return "", err
		}
		if uidExist {
			ticket, err := util.GenerateTokenWithExpireTime(fmt.Sprintf("%s-login", username), model.LOGIN_TICKET_EXPIRE)
			if err != nil {
				return "", err
			}
			// 5min expire
			model.Rdb.Set(ctx, "LOGIN_TICKET:"+username, ticket, model.LOGIN_TICKET_EXPIRE)
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
	nilUser := &model.User{}
	if err != nil {
		return nilUser, err
	}
	username, claimsError := jwtClaims.GetSubject()
	if claimsError != nil {
		return nilUser, claimsError
	}

	rToken, err := model.Rdb.Get(ctx, fmt.Sprintf("TOKEN:%s", username)).Result()
	if err != nil {
		if err == redis.Nil {
			return nilUser, result.AUTH_ERROR
		}
		return nilUser, err
	}

	if rToken == "" || rToken != jwt {
		return nilUser, result.AUTH_ERROR
	}

	return model.UserInfo(username + "test")
	//return model.UserInfo(username)
}

func SendEmail(username string, ticket string) error {
	val, err := model.Rdb.Get(ctx, ticket).Result()
	if err != nil {
		// key does not exists
		if err == redis.Nil {
			return result.CHECK_TICKET_NOTFOUND
		}
		return err
	}

	// Determine if the ticket is correct
	if val != model.REGISTER_STATUS["VERIFY_ACCOUNT"] {
		return result.TICKET_NOT_CORRECT
	}
	code := model.GenerateVerifyCode(username)
	codeKey := "CAPTCHA-" + username
	model.Rdb.Set(ctx, codeKey, code, model.CAPTCHA_EXPIRE_TIME)
	content := model.InsertCode(code)
	emailErr := model.SendEmail(username, content)
	if emailErr != nil {
		return emailErr
	}
	serviceLogger.Infof("Send Email to [%s] with code [%s]\n", username, code)
	// Update the status of the ticket
	model.Rdb.Set(ctx, ticket, model.REGISTER_STATUS["SEND_EMAIL"], model.REGISTER_TICKET_EXPIRE_TIME)
	return nil
}

func CheckVerifyCode(ticket string, code string) error {
	status, err := model.Rdb.Get(ctx, ticket).Result()
	if err != nil {
		if err == redis.Nil {
			return result.CHECK_TICKET_NOTFOUND
		}
		return err
	}
	if status != model.REGISTER_STATUS["SEND_EMAIL"] {
		return result.TICKET_NOT_CORRECT
	}
	username, uErr := util.GetUsername(ticket)
	if uErr != nil {
		return uErr
	}

	codeKey := "CAPTCHA-" + username
	rCode, cErr := model.Rdb.Get(ctx, codeKey).Result()
	if cErr != nil {
		if cErr == redis.Nil {
			return result.CaptchaError
		}
		return cErr
	}

	if code != rCode {
		return result.CaptchaError
	}

	// Update the status of the ticket
	model.Rdb.Set(ctx, ticket, model.REGISTER_STATUS["VERIFY_CAPTCHA"], model.REGISTER_TICKET_EXPIRE_TIME)
	return nil
}

func CheckToken(key string, token string) bool {
	val, err := model.Rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	if val != token {
		return false
	}
	return true
}
