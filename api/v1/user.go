package v1

import (
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var controllerLogger = log.Log

func Register(ctx *gin.Context) {
	// TODO: fill relevant code
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func UserInfo(ctx *gin.Context) {
	if user, err := service.UserInfo(ctx.GetHeader("TOKEN")); err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": user.Uid,
			}).Error(err)
		ctx.JSON(http.StatusOK, result.Failed(
			result.ERROR_GET_USERINFO_FAIL,
			result.GetMsg(result.ERROR_GET_USERINFO_FAIL),
		))
	} else {
		ctx.JSON(http.StatusOK, result.Success(*user))
	}
}

func SendEmail(ctx *gin.Context) {
	username := ctx.Query("username")
	ticket := ctx.GetHeader("TICKET")
	err := service.SendEmail(username, ticket)
	if err != nil {
		if err.Error() == result.GetMsg(result.ERROR_TICKET_NOT_CORRECT) {
			controllerLogger.WithFields(
				logrus.Fields{
					"username": username,
				}).Error(err)
			ctx.JSON(http.StatusOK, result.Failed(
				result.ERROR_TICKET_NOT_CORRECT,
				result.GetMsg(result.ERROR_TICKET_NOT_CORRECT),
			))
		} else {
			ctx.JSON(http.StatusOK, result.Success(nil))
		}
	}
}
