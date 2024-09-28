package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"
)

// USerInfo returns the user information of the current user.
func (s *APIV1Service) UserInfo(c echo.Context) error {
	ctx := c.Request().Context()
	studentID := request.GetUsername(c.Request())
	user, err := s.UserService.UserInfo(ctx, studentID)
	if err != nil {
		log.Errorf("Failed to find user: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalError)
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, response.UserNotFound)
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"email":  user.Email,
		"userId": user.UID,
	}))
}

// Modify paassword.
func (s *APIV1Service) ChangePassword(c echo.Context) error {
	ctx := c.Request().Context()
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}
	// Get password from form
	oldPassword := c.FormValue("oldPassword")
	newPassword := c.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	// Modify password
	err := s.UserService.ModifyPassword(ctx, studentID, oldPassword, newPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.ChangePasswordError)
	}
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()
	// Get Body from request
	newPassword := c.FormValue("newPassword")
	if newPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	cookie, err := c.Cookie(request.ResetPwdTicketSub)
	if err != nil {
		log.Errorf("Get cookie error: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, response.TicketNotFound)
	}
	ticket := cookie.Value

	currentPhase, err := s.Store.Get(ctx, ticket)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.TicketNotFound)
	}
	// Check which phase current in
	switch currentPhase {
	case request.VerifyStatus["VERIFY_ACCOUNT"], request.VerifyStatus["SEND_EMAIL"]:
		return echo.NewHTTPError(http.StatusBadRequest, response.ResetPasswordError)
	case request.VerifyStatus["SUCCESS"]:
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalError)
	case "":
		return echo.NewHTTPError(http.StatusBadRequest, response.TicketNotFound)
	}

	studentID, err := util.IdentityFromToken(ticket, request.ResetPwdTicketSub)
	if err != nil || studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequiredParams)
	}

	if err := s.UserService.ResetPassword(ctx, studentID, newPassword); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.ResetPasswordError)
	}

	// Set VERIFY_STATUS to 3 if successes
	_ = s.Store.Set(ctx, ticket, request.VerifyStatus["SUCCESS"], request.RegisterTicketExp)
	log.Debugf("Reset password success: %s", studentID)
	return c.JSON(http.StatusOK, response.Success(nil))
}
