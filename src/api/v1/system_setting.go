package v1

import (
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/labstack/echo/v4"
)

func (s *APIV1Service) SystemSetting(c echo.Context) error {
	ctx := c.Request().Context()

	settingType := c.Param("settingType")
	if settingType == "" {
		log.Error("The setting type is empty")
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}
	systemSetting, err := s.SysSettingService.GetSysSetting(ctx, settingType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get the system setting")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"setting": systemSetting,
	}))
}

func (s *APIV1Service) UpsetSystemSetting(c echo.Context) error {
	ctx := c.Request().Context()

	settingType := c.Param("settingType")
	if settingType == "" {
		log.Error("The setting type is empty")
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	var settingValue interface{}
	if err := c.Bind(&settingValue); err != nil {
		log.Error("Failed to bind the setting value", err)
		return echo.NewHTTPError(http.StatusBadRequest, response.REQUIRED_PARAMS)
	}

	if err := s.SysSettingService.UpSetSysSetting(ctx, config.TypeFromString(settingType), settingValue); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.INTENAL_ERROR)
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
