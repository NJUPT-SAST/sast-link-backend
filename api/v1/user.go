package v1

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/NJUPT-SAST/sast-link-backend/model"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var controllerLogger = log.Log

func Register(ctx *gin.Context) {
	// get Body from request
	password, passwordFlag := ctx.GetPostForm("password")
	if !passwordFlag {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
		return
	}
	if password == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.PasswordError))
		return
	}

	ticket := ctx.GetHeader("REGISTER-TICKET")

	currentPhase, _ := model.Rdb.Get(ctx, ticket).Result()
	// check which phase current in
	switch currentPhase {
	case model.VERIFY_STATUS["VERIFY_ACCOUNT"], model.VERIFY_STATUS["SEND_EMAIL"]:
		ctx.JSON(http.StatusOK, result.Failed(result.RegisterPhaseError))
		return
	case model.VERIFY_STATUS["SUCCESS"]:
		ctx.JSON(http.StatusOK, result.Failed(result.UserAlreadyExist))
		return
	case "":
		ctx.JSON(http.StatusOK, result.Failed(result.CheckTicketNotfound))
		return
	}

	username, usernameErr := util.IdentityFromToken(ticket, model.REGIST_TICKET_SUB)
	if usernameErr != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(usernameErr)))
		return
	}

	creErr := service.CreateUserAndProfile(username, password)
	if creErr != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(creErr)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
	// set VERIFY_STATUS to 3 if successes
	model.Rdb.Set(ctx, ticket, model.VERIFY_STATUS["SUCCESS"], model.REGISTER_TICKET_EXP)
}

func CheckVerifyCode(ctx *gin.Context) {
	code, codeFlag := ctx.GetPostForm("captcha")
	var ticket, flag string
	if ctx.GetHeader("REGISTER-TICKET") != "" {
		ticket = ctx.GetHeader("REGISTER-TICKET")
		flag = model.REGIST_TICKET_SUB
	} else if ctx.GetHeader("RESETPWD-TICKET") != "" {
		ticket = ctx.GetHeader("RESETPWD-TICKET")
		flag = model.RESETPWD_TICKET_SUB
	} else {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
	}

	if !codeFlag {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
		return
	}

	codeError := service.CheckVerifyCode(ctx, ticket, code, flag)
	if codeError != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(codeError)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func UserInfo(ctx *gin.Context) {
	user, err := service.UserInfo(ctx)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": user.Uid,
			}).Error(err)
		ctx.JSON(http.StatusOK, result.Failed(result.GetUserinfoFail))
		return
	}

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"email":  user.Email,
		"userId": user.Uid,
	}))
}

func SendEmail(ctx *gin.Context) {
	//获取传入TICKET参数
	var ticket, flag string
	if ctx.GetHeader("REGISTER-TICKET") != "" {
		ticket = ctx.GetHeader("REGISTER-TICKET")
		flag = model.REGIST_TICKET_SUB
	} else if ctx.GetHeader("RESETPWD-TICKET") != "" {
		ticket = ctx.GetHeader("RESETPWD-TICKET")
		flag = model.RESETPWD_TICKET_SUB
	} else {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
		return
	}

	username, usernameErr := util.IdentityFromToken(ticket, flag)
	// 错误处理机制写玉玉了
	// 我开始乱写了啊啊啊啊
	if usernameErr != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(usernameErr)
		ctx.JSON(http.StatusUnauthorized, result.Failed(result.TicketNotCorrect))
		return
	}
	// verify if the user email correct
	matched, _ := regexp.MatchString("^[BPFQbpfq](1[7-9]|2[0-9])([0-3])\\d{5}@njupt.edu.cn$", username)
	if !matched {
		ctx.JSON(http.StatusOK, result.Failed(result.UserEmailError))
		return
	}

	var title string
	if flag == model.REGIST_TICKET_SUB {
		title = "确认电子邮件注册SAST-Link账户（无需回复）"
	} else {
		title = "确认电子邮件重置SAST-Link账户密码（无需回复）"
	}
	err := service.SendEmail(ctx, username, ticket, title)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
	} else {
		ctx.JSON(http.StatusOK, result.Success(nil))
	}
}

func VerifyAccount(ctx *gin.Context) {
	username := ctx.Query("username")
	// Capitalize the username
	username = strings.ToLower(username)

	flag := ctx.Query("flag")
	tKey := ""
	// 0 is register
	// 1 is login
	// 2 is resetPassword
	if flag == "0" {
		tKey = model.REGIST_TICKET_SUB
	} else if flag == "1" {
		tKey = model.LOGIN_TICKET_SUB
	} else if flag == "2" {
		tKey = model.RESETPWD_TICKET_SUB
	} else {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
		return
	}

	ticket, err := service.VerifyAccount(ctx, username, flag)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)

		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{tKey: ticket}))
}

func Login(ctx *gin.Context) {
	ticket := ctx.GetHeader("LOGIN-TICKET")
	password := ctx.PostForm("password")
	if ticket == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.CheckTicketNotfound))
		return
	}
	if password == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.PasswordEmpty))
		return
	}

	// Get username from ticket
	username, err := util.IdentityFromToken(ticket, model.LOGIN_TICKET_SUB)
	if err != nil || username == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.TicketNotCorrect))
		return
	}
	//verify if the user is deleted

	// Check the password
	uid, err := service.Login(username, password)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)
		//ctx.JSON(http.StatusUnauthorized, result.Failed(result.VerifyAccountError))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.VerifyPasswordError))
		return
	}
	if uid == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.VerifyAccountError))
		return
	}

	// Set Token with expire time and return
	token, err := util.GenerateTokenWithExp(ctx, model.LoginJWTSubKey(uid), model.LOGIN_TOKEN_EXP)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.GenerateToken))
	}
	model.Rdb.Set(ctx, model.LoginTokenKey(uid), token, model.LOGIN_TOKEN_EXP)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		model.LOGIN_TOKEN_SUB: token,
	}))

	// Oauth: check if need to bound oauth servers like lark, github...
	if oauth_token := ctx.Request.Header["OAUTH-TICKET"][0]; oauth_token != "" {
		audience, err := util.TokenAudience(oauth_token)
		if err != nil {
			ctx.JSON(http.StatusOK, result.Failed(result.OauthTokenError))
			log.Log.Errorln("util.TokenAudience ::: ", err)
			return
		}
		flagIn := strings.Split(audience[0], "-")[1]

		switch flagIn {
		case "oauthLarkToken":
			unionID, err := util.IdentityFromToken(oauth_token, model.OAUTH_LARK_SUB)
			if err != nil {
				ctx.JSON(http.StatusOK, result.Failed(result.OauthTokenError))
				log.Log.Errorln("util.IdentityFromToken ::: ", err)
				return
			}
		    // bind this user with lark union_id and other lark specific user info
			err = service.UpdateLarkUnionID(username, unionID)
			if err != nil {
				ctx.JSON(http.StatusOK, result.Failed(result.OauthTokenError))
				log.Log.Errorln("service.UpdateLarkUnionID ::: ", err)
				return
			}

			// TODO: save oauth user info
			// saveLarkUserInfo()

		default:
			ctx.JSON(http.StatusOK, result.Failed(result.OauthTokenError))
			return
		}
	}
}

// Modify paassword
func ChangePassword(ctx *gin.Context) {
	// Get username from token
	token := ctx.GetHeader("TOKEN")
	uid, err := util.IdentityFromToken(token, model.LOGIN_TOKEN_SUB)
	if err != nil || uid == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.TicketNotCorrect))
		return
	}
	// Get password from form
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if oldPassword == "" || newPassword == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.PasswordEmpty))
		return
	}
	// Modify password
	err = service.ModifyPassword(ctx, uid, oldPassword, newPassword)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func ResetPassword(ctx *gin.Context) {
	// get Body from request
	newPassword, passwordFlag := ctx.GetPostForm("newPassword")
	if !passwordFlag {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
		return
	}
	if newPassword == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.PasswordError))
		return
	}

	ticket := ctx.GetHeader("RESETPWD-TICKET")

	currentPhase, _ := model.Rdb.Get(ctx, ticket).Result()
	// check which phase current in
	switch currentPhase {
	case model.VERIFY_STATUS["VERIFY_ACCOUNT"], model.VERIFY_STATUS["SEND_EMAIL"]:
		ctx.JSON(http.StatusOK, result.Failed(result.ResetPasswordEror))
		return
	case model.VERIFY_STATUS["SUCCESS"]:
		ctx.JSON(http.StatusOK, result.Failed(result.AlreadySetPasswordErr))
		return
	case "":
		ctx.JSON(http.StatusOK, result.Failed(result.CheckTicketNotfound))
		return
	}

	username, usernameErr := util.IdentityFromToken(ticket, model.RESETPWD_TICKET_SUB)
	if usernameErr != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(usernameErr)))
		return
	}

	creErr := service.ResetPassword(username, newPassword)
	if creErr != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(creErr)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))

	// set VERIFY_STATUS to 3 if successes
	model.Rdb.Set(ctx, ticket, model.VERIFY_STATUS["SUCCESS"], model.REGISTER_TICKET_EXP)
}

func Logout(ctx *gin.Context) {
	//verify information
	token := ctx.GetHeader("TOKEN")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}
	//remove Token from uid
	uid, err := util.IdentityFromToken(token, model.LOGIN_TOKEN_SUB)
	if err != nil || uid == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.TokenError))
		return
	}
	model.Rdb.Del(ctx, model.LoginTokenKey(uid))
	ctx.JSON(http.StatusOK, result.Success(nil))
}
