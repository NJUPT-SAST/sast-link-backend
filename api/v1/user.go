package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var controllerLogger = log.Log

func Register(ctx *gin.Context) {
	// TODO: fill relevant code
	// 从ctx中拿到Body
	username, usernameFlag := ctx.GetPostForm("username")
	password, passwordFlag := ctx.GetPostForm("password")
	code, codeFlag := ctx.GetPostForm("code")
	if !usernameFlag || !passwordFlag || !codeFlag {
		ctx.JSON(http.StatusBadRequest, result.ReadBodyError)
		return
	}
	if username == "" || password == "" {
		ctx.JSON(http.StatusBadRequest, result.UsernameOrPasswordError)
		return
	}
	if code == "" {
		ctx.JSON(http.StatusBadRequest, result.VerifyCodeError)
		return
	}

	fmt.Println(username, password, code)

	codeError := service.CheckVerifyCode(username, code)
	if codeError != nil {
		if errors.Is(codeError, result.VerifyCodeError) {
			ctx.JSON(http.StatusBadRequest, result.VerifyCodeError)
			return
		}
	}
	service.CreateUser(username, password)
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func UserInfo(ctx *gin.Context) {
	if user, err := service.UserInfo(ctx.GetHeader("TOKEN")); err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": user.Uid,
			}).Error(err)
		ctx.JSON(http.StatusOK, result.Failed(result.GET_USERINFO_FAIL))
	} else {
		ctx.JSON(http.StatusOK, result.Success(gin.H{
			"email": user.Email,
		}))
	}
}

func SendEmail(ctx *gin.Context) {
	ticket := ctx.GetHeader("TICKET")
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
		if errors.Is(err, result.TICKET_NOT_CORRECT) {
			ctx.JSON(http.StatusUnauthorized, result.Failed(result.TICKET_NOT_CORRECT))
		} else if errors.Is(err, result.CHECK_TICKET_NOTFOUND) {
			ctx.JSON(http.StatusUnauthorized, result.Failed(result.CHECK_TICKET_NOTFOUND))
		} else {
			ctx.JSON(http.StatusUnauthorized, result.Failed(result.SendEmailError))
		}
	} else {
		ctx.JSON(http.StatusOK, result.Success(nil))
	}
}

func VerifyAccount(ctx *gin.Context) {
	username := ctx.Query("username")
	isExist, ticket, err := service.VerifyAccount(username)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)
		//ctx.JSON(http.StatusUnauthorized, result.Failed(result.VerifyAccountError))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.VerifyAccountError))
		return
	}
	if isExist {
		ctx.JSON(http.StatusUnauthorized, result.Failed(result.UserIsExist))
	}
	ctx.JSON(http.StatusOK, result.Success(ticket))
}
