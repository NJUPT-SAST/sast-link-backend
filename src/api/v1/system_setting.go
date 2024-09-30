package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
)

// TODO: Need to authrize the user
func (s *APIV1Service) SystemSetting(c echo.Context) error {
	ctx := c.Request().Context()

	settingType := c.Param("setting_type")
	if settingType == "" {
		log.Error("The setting type is empty")
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}
	systemSetting, err := s.SysSettingService.GetSysSetting(ctx, settingType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get the system setting")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"setting": systemSetting,
	}))
}

// TODO: Need to authrize the user
func (s *APIV1Service) UpsetSystemSetting(c echo.Context) error {
	ctx := c.Request().Context()

	settingType := c.Param("setting_type")
	if settingType == "" {
		log.Error("The setting type is empty")
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	var settingValue interface{}
	if err := c.Bind(&settingValue); err != nil {
		log.Error("Failed to bind the setting value", err)
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	if err := s.SysSettingService.UpSetSysSetting(ctx, config.TypeFromString(settingType), settingValue); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalError)
	}
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) ListIDPName(c echo.Context) error {
	ctx := c.Request().Context()

	idps, err := s.SysSettingService.ListIDPName(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list the identity provider")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"idps": idps,
	}))
}

func (s *APIV1Service) IDPInfo(c echo.Context) error {
	ctx := c.Request().Context()

	idp := c.QueryParam("idp")
	idpInfo, err := s.SysSettingService.IDPInfo(ctx, idp)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get idpInfo")
	}
	return c.JSON(http.StatusOK, response.Success(idpInfo))
}
