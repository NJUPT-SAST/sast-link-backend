package v1

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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
	// TODO: fill relevant code
	// 从ctx中拿到Body
	username, usernameFlag := ctx.GetPostForm("username")
	password, passwordFlag := ctx.GetPostForm("password")
	code, codeFlag := ctx.GetPostForm("code")
	if !usernameFlag || !passwordFlag || !codeFlag {
		ctx.JSON(http.StatusBadRequest, result.ParamError)
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
	// redis ticket is username-register
	username = strings.Split(username, "-")[0]
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
	flag := ctx.Query("flag")
	ticket, err := service.VerifyAccount(username, flag)
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": username,
			}).Error(err)
		if errors.Is(err, result.UserIsExist) {
			ctx.JSON(http.StatusUnauthorized, result.Failed(result.UserIsExist))
		} else if errors.Is(err, result.UserNotExist) {
			ctx.JSON(http.StatusUnauthorized, result.Failed(result.UserNotExist))
		} else if errors.Is(err, result.ParamError) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.ParamError))
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.VerifyAccountError))
		}
		return
	}
	ctx.JSON(http.StatusOK, result.Success(ticket))
}

func Login(ctx *gin.Context) {
	//verify information
	ticket := ctx.GetHeader("LOGIN_TICKET")
	password := ctx.Query("password")
	if ticket == "" {
		ctx.JSON(http.StatusBadRequest, result.AUTH_INCOMING_TICKET_FAIL)
		return
	}
	if password == "" {
		ctx.JSON(http.StatusBadRequest, result.Password_NOTFOUND)
		return
	}
	fmt.Println(ticket, password)
	//get username from ticket
	username, err := util.GetUsername(ticket)
	if err != nil || username == "" {
		ctx.JSON(http.StatusBadRequest, result.TICKET_NOT_CORRECT)
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
	token, err := util.GenerateToken(username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, result.GENERATE_TOKEN)
	}
	model.Rdb.Set(ctx, "TOKEN:"+username, token, time.Hour*6)
	ctx.JSON(http.StatusOK, result.Success(token))
}

func Logout(ctx *gin.Context) {
	//verify information
	token := ctx.GetHeader("TOKEN")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, result.TICKET_NOT_CORRECT)
		return
	}
	fmt.Println(token)
	//remove Token from username
	username, err := util.GetUsername(token)
	if err != nil || username == "" {
		ctx.JSON(http.StatusBadRequest, result.TICKET_NOT_CORRECT)
		return
	}
	model.Rdb.Del(ctx, "TOKEN:"+username)
	ctx.JSON(http.StatusOK, result.Success(nil))
}
