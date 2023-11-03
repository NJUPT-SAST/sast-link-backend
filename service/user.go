package service

import (
	"regexp"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	_ "github.com/redis/go-redis/v9"
)

var serviceLogger = log.Log

// password can just contain ascii character
func CheckPasswordFormat(password string) bool {
	passReg := regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_=+]{6,32}$`)
	return passReg.MatchString(password)
}

func CreateUserAndProfile(email string, password string) error {
	// split email with @
	split := regexp.MustCompile(`@`)
	uid := split.Split(email, 2)[0]
	uid = strings.ToLower(uid)

	if !CheckPasswordFormat(password) {
		return result.PasswordIllegal
	}
	var err error
	//encrypt password
	pwdEncrypt := util.ShaHashing(password)

	err = model.CreateUserAndProfile(&model.User{
		Email:    &email,
		Password: &pwdEncrypt,
		Uid:      &uid,
	}, &model.Profile{
		Nickname: &uid,
		Email:    &email,
		OrgId:    -1,
	})

	if err != nil {
		return err
	} else {
		return nil
	}
}

// In VerifyAccountRegister and VerifyAccountResetPWD, the username must be email
// In VerifyAccountLogin, the username can be email or uid
func VerifyAccount(ctx *gin.Context, username, flag string) (string, error) {
	// 0 is register
	// 1 is login
	// 2 is resetPassword
	if flag == "0" {
		return VerifyAccountRegister(ctx, username)
	} else if flag == "1" {
		return VerifyAccountLogin(ctx, username)
	} else if flag == "2" {
		return VerifyAccountResetPWD(ctx, username)
	} else {
		return "", result.RequestParamError
	}
}

// This username is email
func VerifyAccountResetPWD(ctx *gin.Context, username string) (string, error) {
	// verify if the user email correct
	matched, _ := regexp.MatchString("^[BPFQbpfq](1[7-9]|2[0-9])([0-3])\\d{5}@njupt.edu.cn$", username)
	if !matched {
		return "", result.UserEmailError
	}
	// check if the user is exist
	user, err := model.GetUserByEmail(username)
	if err != nil {
		return "", err
	}

	// User exist and try to reset password
	if user != nil {
		ticket, err := util.GenerateTokenWithExp(ctx, model.ResetPwdJWTSubkey(username), model.RESETPWD_TICKET_EXP)
		if err != nil {
			return "", err
		}
		model.Rdb.Set(ctx, ticket, model.VERIFY_STATUS["VERIFY_ACCOUNT"], model.RESETPWD_TICKET_EXP)
		return ticket, err
	} else {
		// user not exist	and can`t resetPWD
		return "", result.UserNotExist
	}
}

// This function is used to verify the user's email is exist or not when register
// This username is email
func VerifyAccountRegister(ctx *gin.Context, username string) (string, error) {
	// verify if the user email correct
	matched, _ := regexp.MatchString("^[BPFQbpfq](1[7-9]|2[0-9])([0-3])\\d{5}@njupt.edu.cn$", username)
	if !matched {
		return "", result.UserEmailError
	}
	// check if the user is exist
	user, err := model.GetUserByEmail(username)
	if err != nil {
		return "", err
	}
	// user is exist and can't register
	if user != nil {
		return "", result.UserIsExist
	} else {
		// user is not exist and can register
		// generate token and set expire time
		ticket, err := util.GenerateTokenWithExp(ctx, model.RegisterJWTSubKey(username), model.REGISTER_TICKET_EXP)
		if err != nil {
			return "", err
		}
		// set token to redis
		model.Rdb.Set(ctx, ticket, model.VERIFY_STATUS["VERIFY_ACCOUNT"], model.REGISTER_TICKET_EXP)
		return ticket, err
	}
}

// This function is used to verify the user's email is exist or not when login
// This username is email or uid
func VerifyAccountLogin(ctx *gin.Context, username string) (string, error) {
	var user *model.User
	user, err := model.GetUserByEmail(username)
	if err != nil || user == nil {
		return "", result.UserNotExist
	}

	if user == nil {
		user, err := model.GetUserByUid(username)
		if err != nil || user == nil {
			return "", result.UserNotExist
		}
	}

	ticket, err := util.GenerateTokenWithExp(ctx, model.LoginTicketJWTSubKey(*user.Uid), model.LOGIN_TICKET_EXP)
	if err != nil || ticket == "" {
		return "", err
	}
	model.Rdb.Set(ctx, model.LoginTicketKey(*user.Uid), ticket, model.LOGIN_TICKET_EXP)
	return ticket, err
}

func Login(username string, password string) (string, error) {
	// Check password
	uid, err := model.CheckPassword(username, password)
	return uid, err
}

func ModifyPassword(ctx *gin.Context, username, oldPassword, newPassword string) error {
	// Check password
	uid, err := model.CheckPassword(username, oldPassword)
	if err != nil {
		return err
	}

	if uid == "" {
		return result.VerifyAccountError
	}
	pErr := model.ChangePassword(uid, newPassword)
	if pErr != nil {
		return pErr
	}
	return nil
}

func ResetPassword(username, newPassword string) error {
	// Check password form
	if !CheckPasswordFormat(newPassword) {
		return result.PasswordIllegal
	}
	cErr := model.ChangePassword(username, newPassword)
	if cErr != nil {
		return cErr
	}
	return nil
}

func UserInfo(ctx *gin.Context) (*model.User, error) {
	token := ctx.GetHeader("TOKEN")
	nilUser := &model.User{}
	uid, err := util.GetUsername(token, model.LOGIN_TOKEN_SUB)
	if err != nil {
		return nilUser, err
	}

	rToken, err := model.Rdb.Get(ctx, model.LoginTokenKey(uid)).Result()
	if err != nil {
		if err == redis.Nil {
			return nilUser, result.TokenError
		}
		return nilUser, err
	}

	if rToken == "" || rToken != token {
		return nilUser, result.TokenError
	}

	return model.UserInfo(uid)
}

func SendEmail(ctx *gin.Context, username, ticket, title string) error {
	val, err := model.Rdb.Get(ctx, ticket).Result()
	if err != nil {
		// key does not exists
		if err == redis.Nil {
			return result.CheckTicketNotfound
		}
		return err
	}

	// Determine if the ticket is correct
	if val != model.VERIFY_STATUS["VERIFY_ACCOUNT"] {
		return result.TicketNotCorrect
	}
	code := model.GenerateVerifyCode()
	model.Rdb.Set(ctx, model.CaptchaKey(username), code, model.CAPTCHA_EXP)
	content := model.InsertCode(code)
	emailErr := model.SendEmail(username, content, title)
	if emailErr != nil {
		return emailErr
	}
	serviceLogger.Infof("Send Email to [%s] with code [%s]\n", username, code)
	// Update the status of the ticket
	model.Rdb.Set(ctx, ticket, model.VERIFY_STATUS["SEND_EMAIL"], model.REGISTER_TICKET_EXP)
	return nil
}

func CheckVerifyCode(ctx *gin.Context, ticket, code, flag string) error {
	status, err := model.Rdb.Get(ctx, ticket).Result()
	if err != nil {
		if err == redis.Nil {
			return result.CheckTicketNotfound
		}
		return err
	}
	if status != model.VERIFY_STATUS["SEND_EMAIL"] {
		return result.TicketNotCorrect
	}
	username, uErr := util.GetUsername(ticket, flag)
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
	model.Rdb.Set(ctx, ticket, model.VERIFY_STATUS["VERIFY_CAPTCHA"], model.REGISTER_TICKET_EXP)
	return nil
}

func UpdateUserGitHubId(username, githubId string) error {
	return model.UpdateGithubId(username, githubId)
}

func GetUserByGithubId(githubId string) (*model.User, error) {
	return model.FindUserByGithubId(githubId)
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
