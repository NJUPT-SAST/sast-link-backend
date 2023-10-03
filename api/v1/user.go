package v1

import (
	"net/http"
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

	value, _ := ctx.Get("username")
	username := value.(string)
	creErr := service.CreateUser(username, password)
	if creErr != nil {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(creErr)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))

	// set REGISTER_STATUS to 3 if successes
	model.Rdb.Set(ctx, username, model.REGISTER_STATUS["SUCCESS"], model.REGISTER_TICKET_EXP)
}

func CheckVerifyCode(ctx *gin.Context) {
	code, codeFlag := ctx.GetPostForm("captcha")
	ticket, _ := ctx.Get("token")
	if !codeFlag {
		ctx.JSON(http.StatusOK, result.Failed(result.RequestParamError))
		return
	}

	codeError := service.CheckVerifyCode(ctx, ticket.(string), code)
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
	ticket, _ := ctx.Get("token")
	value, _ := ctx.Get("username")
	username := value.(string)
	// 错误处理机制写玉玉了
	// 我开始乱写了啊啊啊啊
	//if exists == false {
	//	controllerLogger.WithFields(
	//		logrus.Fields{
	//			"username": username,
	//		}).Error(result.UserNameEmpty)
	//	ctx.JSON(http.StatusUnauthorized, result.Failed(result.TicketNotCorrect))
	//	return
	//}
	err := service.SendEmail(ctx, username, ticket.(string))
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
	if flag == "1" {
		tKey = "login_ticket"
	} else if flag == "0" {
		tKey = "register_ticket"
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
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		tKey: ticket,
	}))
}

func Login(ctx *gin.Context) {
	password := ctx.PostForm("password")
	if password == "" {
		ctx.JSON(http.StatusOK, result.Failed(result.PasswordEmpty))
		return
	}

	// Get username from gin.Context
	value, _ := ctx.Get("username")
	username := value.(string)
	// Check the password
	err := service.Login(username, password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(err)))
		return
	}
	// Set Token with expire time and return
	// TODO: role
	token, err := util.GenerateTokenWithExp(model.LoginJWTSubKey(username), model.LOGIN_TOKEN_SUB, 0, model.LOGIN_TOKEN_EXP)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Failed(result.GenerateToken))
	}
	model.Rdb.Set(ctx, model.LoginTokenKey(username), token, model.LOGIN_TOKEN_EXP)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"token": token,
	}))
}

// Modify paassword
func ChangePassword(ctx *gin.Context) {
	// Get username from token
	username, exists := ctx.Get("username")
	if exists == false {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.AuthError))
		return
	}
	// Get password from form
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if oldPassword == "" || newPassword == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.PasswordEmpty))
		return
	}
	// Modify password
	err := service.ModifyPassword(ctx, username.(string), oldPassword, newPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(err)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func Logout(ctx *gin.Context) {
	//verify information
	token := ctx.GetHeader("TOKEN")
	if token == "" {
		ctx.JSON(http.StatusOK, result.AuthIncomingTokenFail)
		return
	}
	//remove Token from username
	username, exists := ctx.Get("username")
	if exists == false {
		ctx.JSON(http.StatusOK, result.AuthError)
		return
	}
	model.Rdb.Del(ctx, model.LoginTokenKey(username.(string)))
	ctx.JSON(http.StatusOK, result.Success(nil))
}
