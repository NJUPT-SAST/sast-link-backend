package v1

import (
	"net/http"
	"regexp"
	"time"

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
		ctx.JSON(http.StatusBadRequest, result.Failed(result.ParamError))
		return
	}
	if password == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.UsernameOrPasswordError))
		return
	}

	ticket := ctx.GetHeader("REGISTER_TICKET")
	username, usernameErr := util.GetUsername(ticket)
	if usernameErr != nil {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(usernameErr)))
		return
	}

	creErr := service.CreateUser(username, password)
	if creErr != nil {
		ctx.JSON(http.StatusBadRequest, result.UnknownError)
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func CheckVerifyCode(ctx *gin.Context) {
	code, codeFlag := ctx.GetPostForm("captcha")
	ticket := ctx.GetHeader("REGISTER_TICKET")
	if !codeFlag {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.ParamError))
		return
	}

	codeError := service.CheckVerifyCode(ticket, code)
	if codeError != nil {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(codeError)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func UserInfo(ctx *gin.Context) {
	user, err := service.UserInfo(ctx.GetHeader("TOKEN"))
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": user.Uid,
			}).Error(err)
		ctx.JSON(http.StatusOK, result.Failed(result.GET_USERINFO_FAIL))
		return
	}

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"email": user.Email,
	}))
}

func SendEmail(ctx *gin.Context) {
	ticket := ctx.GetHeader("REGISTER_TICKET")
	username, usernameErr := util.GetUsername(ticket)
	// 错误处理机制写玉玉了
	// 我开始乱写了啊啊啊啊
	if usernameErr != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(usernameErr)
		ctx.JSON(http.StatusUnauthorized, result.Failed(result.TICKET_NOT_CORRECT))
		return
	}

	err := service.SendEmail(username, ticket)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)

		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(err)))
	} else {
		ctx.JSON(http.StatusOK, result.Success(nil))
	}
}

func VerifyAccount(ctx *gin.Context) {
	username := ctx.Query("username")
	flag := ctx.Query("flag")
	tKey := ""
	// 0 is register
	// 1 is login
	if flag == "1" {
		tKey = "login_ticket"
	} else if flag == "0" {
		tKey = "register_ticket"
	} else {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.ParamError))
		return
	}
	ticket, err := service.VerifyAccount(username, flag)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)

		ctx.JSON(http.StatusBadRequest, result.Failed(result.HandleError(err)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		tKey: ticket,
	}))
}

func Login(ctx *gin.Context) {
	//verify information
	ticket := ctx.GetHeader("LOGIN_TICKET")
	password := ctx.PostForm("password")
	if ticket == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.CHECK_TICKET_NOTFOUND))
		return
	}
	if password == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.Password_NOTFOUND))
		return
	}
	//get username from ticket
	username, err := util.GetUsername(ticket)
	if err != nil || username == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.TICKET_NOT_CORRECT))
		return
	}
	//transform username
	compile, err := regexp.Compile("-")
	split := compile.Split(username, 2)
	username = split[0]
	//check the password
	flag, err := service.Login(username, password)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)
		//ctx.JSON(http.StatusUnauthorized, result.Failed(result.VerifyAccountError))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.VerifyPasswordError))
		return
	}
	if !flag {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.PasswordError))
		return
	}
	//set Token with expire time and return
	token, err := util.GenerateTokenWithExpireTime(username, time.Hour*24)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.GENERATE_TOKEN))
	}
	model.Rdb.Set(ctx, "TOKEN:"+username, token, time.Hour*24)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"token": token,
	}))
}

func Logout(ctx *gin.Context) {
	//verify information
	token := ctx.GetHeader("TOKEN")

	if token == "" {
		ctx.JSON(http.StatusBadRequest, result.TICKET_NOT_CORRECT)
		return
	}
	//remove Token from username
	username, err := util.GetUsername(token)
	if err != nil || username == "" {
		ctx.JSON(http.StatusBadRequest, result.TICKET_NOT_CORRECT)
		return
	}
	model.Rdb.Del(ctx, "TOKEN:"+username)
	ctx.JSON(http.StatusOK, result.Success(nil))
}
