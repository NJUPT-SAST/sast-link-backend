package v1

import (
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"regexp"
	"strings"
)

func GetProfile(ctx *gin.Context) {
	token := ctx.GetHeader("TOKEN")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}
	username, err := util.GetUsername(token, model.LOGIN_TOKEN_SUB)
	if username == "" || err != nil {
		controllerLogger.Errorln("Can`t get username by token", err)
		ctx.JSON(http.StatusOK, result.Failed(result.TokenError))
		return
	}
	// split email with @
	split := regexp.MustCompile(`@`)
	uid := split.Split(username, 2)[0]
	uid = strings.ToLower(uid)

	profileInfo, serErr := service.GetProfileInfo(uid)
	if serErr != nil {
		controllerLogger.Errorln("GetProfile service wrong", serErr)
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(serErr)))
		return
	}

	if dep, org, getOrgErr := service.GetProfileOrg(profileInfo.OrgId); getOrgErr != nil {
		controllerLogger.Errorln("GetProfileOrg Err", getOrgErr)
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(getOrgErr)))
		return
	} else {
		ctx.JSON(http.StatusOK, result.Success(gin.H{
			"nickname": profileInfo.Nickname,
			"dep":      dep,
			"org":      org,
			"email":    profileInfo.Email,
			"avatar":   profileInfo.Avatar,
			"bio":      profileInfo.Bio,
			"link":     profileInfo.Link,
			"badge":    profileInfo.Badge,
			"hide":     profileInfo.Hide,
		}))
		return
	}
}
func ChangeProfile(ctx *gin.Context) {
	token := ctx.GetHeader("TOKEN")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}

	username, err := util.GetUsername(token, model.LOGIN_TOKEN_SUB)
	if username == "" || err != nil {
		controllerLogger.Errorln("Can`t get username by token", err)
		ctx.JSON(http.StatusOK, result.Failed(result.TokenError))
		return
	}
	// split email with @
	split := regexp.MustCompile(`@`)
	uid := split.Split(username, 2)[0]
	uid = strings.ToLower(uid)

	//get profile info from body
	profile := model.Profile{}
	if err = ctx.ShouldBindBodyWith(&profile, binding.JSON); err != nil {
		controllerLogger.Errorln("get profile from request body wrong", err)
		ctx.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}

	if serviceErr := service.ChangeProfile(&profile, uid); serviceErr != nil {
		controllerLogger.Errorln("ChangeProfile service wrong", serviceErr)
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(serviceErr)))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(nil))
}

func UploadAvatar(ctx *gin.Context) {
	token := ctx.GetHeader("TOKEN")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}

	username, err := util.GetUsername(token, model.LOGIN_TOKEN_SUB)
	if username == "" || err != nil {
		controllerLogger.Errorln("Can`t get username by token", err)
		ctx.JSON(http.StatusOK, result.Failed(result.TokenError))
		return
	}
	// split email with @
	split := regexp.MustCompile(`@`)
	uid := split.Split(username, 2)[0]
	uid = strings.ToLower(uid)

	//obtain avatar file from body
	avatar, err := ctx.FormFile("avatarFile")
	if err != nil || avatar == nil {
		controllerLogger.Errorln("get avatarFile Error", err)
		ctx.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}

	if err := service.UploadAvatar(avatar, uid, ctx); err != nil {
		controllerLogger.Errorln("uploadAvatar Error", err)
		ctx.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}

	ctx.JSON(http.StatusOK, result.Success(nil))
}
func ChangeEmail(ctx *gin.Context) {
	ctx.JSON(200, "success")
}

func DealCensorRes(ctx *gin.Context) {

}
