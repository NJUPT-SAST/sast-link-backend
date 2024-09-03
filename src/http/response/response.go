package response

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Error return error response for echo
//
// It can handle different types of errors:
// 1. LocalError, if the error is LocalError, it will return the error code and message
// 2. error, if the error is error, it will return the error message
// 3. string, if the error is string, it will return the error message
// if the error is other type, it will return "未知错误"
func Error(c echo.Context, err interface{}) error {
	code := http.StatusInternalServerError
	switch e := err.(type) {
	case LocalError:
		if checkStatusCode(e.ErrCode) {
			code = e.ErrCode
		}
		return c.JSON(code, Failed(e))
	case error:
		return c.JSON(http.StatusInternalServerError, Failed(errors.New(e.Error())))
	case string:
		return c.JSON(http.StatusInternalServerError, Failed(errors.New(e)))
	default:
		return c.JSON(http.StatusInternalServerError, Failed(errors.New("未知错误")))
	}
}

// 100-599 are standard HTTP status codes
func checkStatusCode(code int) bool {
	return code >= 100 && code <= 599
}

type Response struct {
	Success bool        `json:"success"`
	ErrCode int         `json:"err_code"`
	ErrMsg  string      `json:"err_msg"`
	Data    interface{} `json:"data"`
}

func Success(data any) Response {
	return Response{
		Success: true,
		ErrCode: http.StatusOK,
		ErrMsg:  "",
		Data:    data,
	}
}

func Failed(e error) Response {
	if e, ok := e.(LocalError); ok {
		return Response{
			Success: false,
			ErrCode: e.ErrCode,
			ErrMsg:  e.ErrMsg,
			Data:    nil,
		}
	}
	return Response{
		Success: false,
		ErrCode: 5001,
		ErrMsg:  e.Error(),
		Data:    nil,
	}
}

type LocalError struct {
	ErrCode int
	ErrMsg  string
	Err     error
}

// Error implement error interface
func (e LocalError) Error() string {
	return fmt.Sprintf("err_code: %d, err_msg: %s, err: %v", e.ErrCode, e.ErrMsg, e.Err)
}

// Create common error
var (
	// Authorization error
	TICKET_NOT_FOUND   = LocalError{ErrCode: 1001, ErrMsg: "ticket not found"}
	TICKET_INVALID     = LocalError{ErrCode: 1002, ErrMsg: "ticket invalid"}
	PASSWORD_INCORRECT = LocalError{ErrCode: 1003, ErrMsg: "password incorrect"}
	USER_NOT_FOUND     = LocalError{ErrCode: 1004, ErrMsg: "user not found"}
	LOGIN_FAILED       = LocalError{ErrCode: 1005, ErrMsg: "login failed, please check your username and password"}
	// Unauthorized http status code is 401
	UNAUTHORIZED = LocalError{ErrCode: 401, ErrMsg: "unauthorized"}
	// Request error
	REQUIRED_PARAMS = LocalError{ErrCode: 2001, ErrMsg: "required params"}
	// User error
	EMAIL_INVALID         = LocalError{ErrCode: 3001, ErrMsg: "email invalid"}
	USER_EXIST            = LocalError{ErrCode: 3002, ErrMsg: "user exist"}
	VERIFY_CODE_INCORRECT = LocalError{ErrCode: 3003, ErrMsg: "verify code incorrect"}
	CHANGE_PASSWORD_ERROR = LocalError{ErrCode: 3004, ErrMsg: "change password error"}
	RESET_PASSWORD_ERROR  = LocalError{ErrCode: 3005, ErrMsg: "reset password error"}
	// Internal error
	INTENAL_ERROR = LocalError{ErrCode: 5001, ErrMsg: "internal error"}
	// OAuth2 server error
	RESHRESH_TOKEN_INVALID = LocalError{ErrCode: 6001, ErrMsg: "refresh token invalid"}
	CLIENT_ID_INVALID      = LocalError{ErrCode: 6002, ErrMsg: "client id invalid"}
	CLIENT_SECRET_INVALID  = LocalError{ErrCode: 6004, ErrMsg: "client secret invalid"}
	CLIENT_ERROR           = LocalError{ErrCode: 6003, ErrMsg: "client error"}
	CODE_INVALID           = LocalError{ErrCode: 6002, ErrMsg: "code invalid"}
	// Profile error
	PROFILE_INFO_ERROR = LocalError{ErrCode: 7001, ErrMsg: "profile error"}
	PROFILE_ORG_ERROR  = LocalError{ErrCode: 7002, ErrMsg: "profile organization error"}
)

// warp error
func (e *LocalError) Wrap(err error) LocalError {
	e.Err = err
	return *e
}

// determine whether the error is equal
func (e *LocalError) Is(err error) bool {
	if err, ok := err.(LocalError); ok {
		return err.ErrCode == e.ErrCode
	}
	return false
}

// SetCookie sets a cookie
func SetCookie(ctx echo.Context, key, value string) {
	ctx.SetCookie(&http.Cookie{
		Name:     key,
		Value:    value,
		HttpOnly: true,
	})
}

// SetCookieWithExpire sets a cookie with an expiration time
//
// If expire is -1, the cookie will be deleted
func SetCookieWithExpire(ctx echo.Context, key, value string, expire time.Duration) {
	maxAge := int(expire.Seconds())

	secure := false
	sameSite := http.SameSiteStrictMode

	// Determine if the request is HTTPS, if it is, set the cookie to secure and SameSite=None
	// This is to prevent CSRF attacks
	origin := ctx.Response().Header().Get("Origin")
	if origin != "" {
		isHTTPS := strings.HasPrefix(origin, "https://")
		if isHTTPS {
			secure = true
			sameSite = http.SameSiteNoneMode
		}
	}

	ctx.SetCookie(&http.Cookie{
		Name:     key,
		Value:    value,
		HttpOnly: true,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   secure,
		SameSite: sameSite,
	})
}
