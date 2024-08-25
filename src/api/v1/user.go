package v1

import (
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/labstack/echo/v4"
)

// USerInfo returns the user information of the current user
func (s *APIV1Service) UserInfo(c echo.Context) error {
	ctx := c.Request().Context()
	studentID := request.GetUsername(c.Request())
	user, err := s.UserService.UserInfo(ctx, studentID)
	if err != nil {
		log.Errorf("User token error: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	mapUser := map[string]interface{}{
		"email":  user.Email,
		"userId": user.Uid,
	}
	return c.JSON(http.StatusOK, response.Success(mapUser))
}

// Modify paassword
func (s *APIV1Service) ChangePassword(c echo.Context) error {
	ctx := c.Request().Context()
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}
	// Get password from form
	oldPassword := c.FormValue("oldPassword")
	newPassword := c.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	// Modify password
	err := s.UserService.ModifyPassword(ctx, studentID, oldPassword, newPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()
	// Get Body from request
	newPassword := c.FormValue("newPassword")
	if newPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	cookie, err := c.Cookie(request.RESETPWD_TICKET_SUB)
	if err != nil {
		log.Errorf("Get cookie error: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}
	ticket := cookie.Value

	currentPhase, err := s.Store.Get(ctx, ticket)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}
	// Check which phase current in
	switch currentPhase {
	case request.VERIFY_STATUS["VERIFY_ACCOUNT"], request.VERIFY_STATUS["SEND_EMAIL"]:
		return echo.NewHTTPError(http.StatusBadRequest, response.ResetPasswordEror)
	case request.VERIFY_STATUS["SUCCESS"]:
		return echo.NewHTTPError(http.StatusBadRequest, response.AlreadySetPasswordErr)
	case "":
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}

	studentID, err := util.IdentityFromToken(ticket, request.RESETPWD_TICKET_SUB, s.Config.Secret)
	if err != nil || studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	if err := s.UserService.ResetPassword(studentID, newPassword); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	// Set VERIFY_STATUS to 3 if successes
	s.Store.Set(ctx, ticket, request.VERIFY_STATUS["SUCCESS"], request.REGISTER_TICKET_EXP)
	log.Debugf("Reset password success: %s", studentID)
	return c.JSON(http.StatusOK, response.Success(nil))
}
