package v1

import (
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/labstack/echo/v4"
)

func (s *APIV1Service) SystemSetting(c echo.Context) error {
	ctx := c.Request().Context()

	settingType := c.Param("settingType")
	if settingType == "" {
		log.Error("The setting type is empty")
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}
	systemSetting, err := s.SysSettingService.GetSysSetting(ctx, config.TypeFromString(settingType))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	return c.JSON(http.StatusOK, response.Success(systemSetting))
}

func (s *APIV1Service) UpsetSystemSetting(c echo.Context) error {
	ctx := c.Request().Context()

	settingType := c.Param("settingType")
	if settingType == "" {
		log.Error("The setting type is empty")
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	var settingValue interface{}
	if err := c.Bind(&settingValue); err != nil {
		log.Error("Failed to bind the setting value", err)
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	if err := s.SysSettingService.UpSetSysSetting(ctx, config.TypeFromString(settingType), settingValue); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	return c.JSON(http.StatusOK, response.Success(nil))
}
