package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/store"
)

func (s *APIV1Service) GetProfile(c echo.Context) error {
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	profileInfo, err := s.ProfileService.GetProfileInfo(studentID)
	if err != nil {
		s.ProfileLog.Error("Failed to get profile info", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, response.ProfileInfoError)
	}
	dep, org, err := s.GetProfileOrg(profileInfo.OrgID)
	if err != nil {
		s.ProfileLog.Error("Failed to get profile orgenization", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, response.ProfileOrgError)
	}
	profileMap := map[string]interface{}{
		"nickname": profileInfo.Nickname,
		"dep":      dep,
		"org":      org,
		"email":    profileInfo.Email,
		"avatar":   profileInfo.Avatar,
		"bio":      profileInfo.Bio,
		"link":     profileInfo.Link,
		"badge":    profileInfo.Badge,
		"hide":     profileInfo.Hide,
	}
	return c.JSON(http.StatusOK, response.Success(profileMap))
}

func (s *APIV1Service) ChangeProfile(c echo.Context) error {
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	// Get profile info from body
	profile := store.Profile{}
	if err := c.Bind(&profile); err != nil {
		s.ProfileLog.Error("Failed to bind profile info", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	if serviceErr := s.ProfileService.ChangeProfile(&profile, studentID); serviceErr != nil {
		s.ProfileLog.Error("Failed to change profile info", zap.Error(serviceErr))
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalError)
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) UploadAvatar(c echo.Context) error {
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	ctx := c.Request().Context()

	// Obtain avatar file from body
	avatar, err := c.FormFile("avatarFile")
	if err != nil || avatar == nil {
		s.ProfileLog.Error("Failed to get avatar file", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	filePath, uploadSerErr := s.ProfileService.UploadAvatar(ctx, avatar, studentID)
	if uploadSerErr != nil {
		s.ProfileLog.Error("Failed to upload avatar", zap.Error(uploadSerErr))
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalError)
	}

	return c.JSON(http.StatusOK, response.Success(map[string]string{"avatar": filePath}))
}

// TODO: Implement the following functions.
func ChangeEmail(ctx *gin.Context) {
	ctx.JSON(200, "success")
}

// DealCensorRes is a fallback function for the COS callback interface.
func (s *APIV1Service) DealCensorRes(c echo.Context) error {
	ctx := c.Request().Context()
	if header := c.Response().Header().Get("X-Ci-Content-Version"); header != "Simple" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	checkRes := store.CheckRes{}
	if err := c.Bind(&checkRes); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	// Judge if picture review fail or need manual re-review and sent to feishu bot
	sentMsgErr := s.ProfileService.SentMsgToBot(ctx, &checkRes)
	if sentMsgErr != nil {
		s.ProfileLog.Error("Failed to send message to bot", zap.Error(sentMsgErr))
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalError)
	}

	// Mv frozen image and replace database url
	if checkRes.Data.ForbiddenStatus == 1 {
		if err := s.DealWithFrozenImage(ctx, &checkRes); err != nil {
			s.ProfileLog.Error("Failed to deal with frozen image", zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, response.InternalError)
		}
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// BindStatus return the third-party binding status of the user.
func (s *APIV1Service) BindStatus(c echo.Context) error {
	ctx := c.Request().Context()

	stuID := request.GetUsername(c.Request())

	bindList, err := s.GetBindList(ctx, stuID)
	if err != nil {
		s.ProfileLog.Error("Failed to get bind list", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	return c.JSON(http.StatusOK, response.Success(bindList))
}
