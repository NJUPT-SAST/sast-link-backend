package service

import (
	"regexp"
	"strings"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var serviceLogger = log.Log

func CreateUser(email string, password string) error {
	// split email with @
	split := regexp.MustCompile(`@`)
	uid := split.Split(email, 2)[0]
	uid = strings.ToLower(uid)

	// password can just contain ascii character
	passReg := regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_=+]{6,32}$`)
	if !passReg.MatchString(password) {
		return result.PasswordIllegal
	}

	//encrypt password
	pwdEncrypt := util.ShaHashing(password)
	err := model.CreateUser(&model.User{
		Email:    &email,
		Password: &pwdEncrypt,
		Uid:      &uid,
	})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func VerifyAccount(ctx *gin.Context, username, flag string) (string, error) {
	// 0 is register
	// 1 is login
	if flag == "0" {
		return VerifyAccountRegister(ctx, username)
	} else {
		return VerifyAccountLogin(ctx, username)
	}
}

// This function is used to verify the user's email is exist or not when register
func VerifyAccountRegister(ctx *gin.Context, username string) (string, error) {
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
		ticket, err := util.GenerateTokenWithExp(model.RegisterJWTSubKey(username), model.REGISTER_TICKET_EXP)
		if err != nil {
			return "", err
		}
		// set token to redis
		model.Rdb.Set(ctx, ticket, model.REGISTER_STATUS["VERIFY_ACCOUNT"], model.REGISTER_TICKET_EXP)
		return ticket, err
	}
}

// This function is used to verify the user's email is exist or not when login
func VerifyAccountLogin(ctx *gin.Context, username string) (string, error) {
	exist, err := model.CheckUserByEmail(username)
	if err != nil {
		return "", err
	}
	// user is existed and can login
	if exist {
		ticket, err := util.GenerateTokenWithExp(model.LoginJWTSubKey(username), model.LOGIN_TICKET_EXP)
		if err != nil {
			return "", err
		}
		// 5min expire
		model.Rdb.Set(ctx, model.LoginTicketKey(username), ticket, model.LOGIN_TICKET_EXP)
		return ticket, err
	} else { // user is not exist and can't login
		// login can use uid and email
		uidExist, err := model.CheckUserByUid(username)
		if err != nil {
			return "", err
		}
		if uidExist {
			ticket, err := util.GenerateTokenWithExp(model.LoginJWTSubKey(username), model.LOGIN_TICKET_EXP)
			if err != nil {
				return "", err
			}
			// 5min expire
			model.Rdb.Set(ctx, model.LoginTicketKey(username), ticket, model.LOGIN_TICKET_EXP)
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

func UserInfo(ctx *gin.Context) (*model.User, error) {
	token := ctx.GetHeader("TOKEN")
	nilUser := &model.User{}
	username, err := util.GetUsername(token)
	if err != nil {
		return nilUser, err
	}

	rToken, err := model.Rdb.Get(ctx, model.LoginTokenKey(username)).Result()
	if err != nil {
		if err == redis.Nil {
			return nilUser, result.AuthError
		}
		return nilUser, err
	}

	if rToken == "" || rToken != token {
		return nilUser, result.AuthError
	}

	return model.UserInfo(username)
}

func SendEmail(ctx *gin.Context, username, ticket string) error {
	val, err := model.Rdb.Get(ctx, ticket).Result()
	if err != nil {
		// key does not exists
		if err == redis.Nil {
			return result.CheckTicketNotfound
		}
		return err
	}

	// Determine if the ticket is correct
	if val != model.REGISTER_STATUS["VERIFY_ACCOUNT"] {
		return result.TicketNotCorrect
	}
	code := model.GenerateVerifyCode()
	model.Rdb.Set(ctx, model.CaptchaKey(username), code, model.CAPTCHA_EXP)
	content := model.InsertCode(code)
	emailErr := model.SendEmail(username, content)
	if emailErr != nil {
		return emailErr
	}
	serviceLogger.Infof("Send Email to [%s] with code [%s]\n", username, code)
	// Update the status of the ticket
	model.Rdb.Set(ctx, ticket, model.REGISTER_STATUS["SEND_EMAIL"], model.REGISTER_TICKET_EXP)
	return nil
}

func CheckVerifyCode(ctx *gin.Context, ticket, code string) error {
	status, err := model.Rdb.Get(ctx, ticket).Result()
	if err != nil {
		if err == redis.Nil {
			return result.CheckTicketNotfound
		}
		return err
	}
	if status != model.REGISTER_STATUS["SEND_EMAIL"] {
		return result.TicketNotCorrect
	}
	username, uErr := util.GetUsername(ticket)
	if uErr != nil {
		return uErr
	}

	rCode, cErr := model.Rdb.Get(ctx, model.CaptchaKey(username)).Result()
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
	model.Rdb.Set(ctx, ticket, model.REGISTER_STATUS["VERIFY_CAPTCHA"], model.REGISTER_TICKET_EXP)
	return nil
}

func CheckToken(ctx *gin.Context, key, token string) bool {
	val, err := model.Rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	if val != token {
		return false
	}
	return true
}
