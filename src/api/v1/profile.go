package v1

import (
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func (s *APIV1Service) GetProfile(c echo.Context) error {
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	profileInfo, serErr := s.ProfileService.GetProfileInfo(studentID)
	if serErr != nil {
		log.Errorf("GetProfile service wrong: %s", serErr.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.PROFILE_INFO_ERROR)
	}
	dep, org, getOrgErr := s.GetProfileOrg(profileInfo.OrgId)
	if getOrgErr != nil {
		log.Errorf("GetProfileOrg service wrong: %s", getOrgErr.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.PROFILE_ORG_ERROR)
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
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	// Get profile info from body
	profile := store.Profile{}
	if err := c.Bind(&profile); err != nil {
		log.Errorf("Bind profile error: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	if serviceErr := s.ProfileService.ChangeProfile(&profile, studentID); serviceErr != nil {
		log.Errorf("ChangeProfile service wrong: %s", serviceErr.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.INTENAL_ERROR)
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) UploadAvatar(c echo.Context) error {
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	ctx := c.Request().Context()

	// Obtain avatar file from body
	avatar, err := c.FormFile("avatarFile")
	if err != nil || avatar == nil {
		log.Errorf("Get avatar file error: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	filePath, uploadSerErr := s.ProfileService.UploadAvatar(avatar, studentID, ctx)
	if uploadSerErr != nil {
		log.Errorf("UploadAvatar service wrong: %s", uploadSerErr.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.INTENAL_ERROR)
	}

	return c.JSON(http.StatusOK, response.Success(map[string]string{"avatar": filePath}))
}

// TODO: Implement the following functions
func ChangeEmail(ctx *gin.Context) {
	ctx.JSON(200, "success")
}

// DealCensorRes is a fallback function for the COS callback interface.
func (s *APIV1Service) DealCensorRes(c echo.Context) error {
	ctx := c.Request().Context()
	if header := c.Response().Header().Get("X-Ci-Content-Version"); header != "Simple" {
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	checkRes := store.CheckRes{}
	if err := c.Bind(&checkRes); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	// Judge if picture review fail or need manual re-review and sent to feishu bot
	sentMsgErr := s.ProfileService.SentMsgToBot(ctx, &checkRes)
	if sentMsgErr != nil {
		log.Errorf("SentMsgToBot service wrong: %s", sentMsgErr.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.INTENAL_ERROR)
	}

	// Mv frozen image and replace database url
	if checkRes.Data.ForbiddenStatus == 1 {
		if err := s.DealWithFrozenImage(ctx, &checkRes); err != nil {
			log.Errorf("DealWithFrozenImage service wrong: %s", err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, response.INTENAL_ERROR)
		}
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}
