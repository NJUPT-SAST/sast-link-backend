package v1

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/plugin/idp/oauth2"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/NJUPT-SAST/sast-link-backend/validator"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
)

func (s *APIV1Service) Login(c echo.Context) error {
	// LOGIN-TICKET just save username etc...
	// not used for login authentication
	// confused though...
	ctx := c.Request().Context()

	password := c.FormValue("password")
	if password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	username := request.GetUsername(c.Request())

	if username == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, response.TicketNotCorrect)
	}

	uid, err := s.UserService.Login(username, password)
	if err != nil {
		log.Errorf("Login fail: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	if uid == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, response.VerifyAccountError)
	}

	// Set Token with expire time and return
	token, err := util.GenerateTokenWithExp(ctx, request.LoginJWTSubKey(uid), s.Config.Secret, request.LOGIN_ACCESS_TOKEN_EXP)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	s.Store.Set(ctx, request.LoginTokenKey(uid), token, request.LOGIN_ACCESS_TOKEN_EXP)
	response.SetCookie(c, request.AccessTokenCookieName, token)

	return c.JSON(http.StatusOK, response.Success(token))
}

// LoginWithSSO login with SSO, it will exchange the token with the SSO identity provider and get the user info.
//
// If the user is not registered, it will redirect to the frontend to bind the email.
func (s *APIV1Service) LoginWithSSO(c echo.Context) error {
	ctx := c.Request().Context()
	// Get Idp name from query
	idpName := c.QueryParam("idp")
	identityProvider, err := s.Store.GetIdentityProviderByName(idpName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}
	if identityProvider == nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}

	var userInfo *oauth2.IdentityProviderUserInfo
	if identityProvider.Type == oauth2.IDPTypeOAuth2 {
		oauth2Idp, err := oauth2.NewOauth2IdentityProvider(identityProvider)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
		}
		token, err := oauth2Idp.ExchangeToken(ctx, identityProvider.GetOauth2Setting(), c.QueryParam("redirect_url"), c.QueryParam("code"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
		}
		userInfo, err = oauth2Idp.UserInfo(ctx, identityProvider.GetOauth2Setting(), token)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
		}
	} else {
		// Now only support OAuth2
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}

	// Get user from our database
	user, err := s.Store.OauthInfoByUID(idpName, userInfo.IdentifierID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}
	studentID := request.GetUsername(c.Request())
	// TODO: Get user info from the sso identity provider.
	if user == nil {
		// User not found, Need to redirect to front end to bind email
		// s.UpsetOauthInfo(studentID, store.LARK_CLIENT_TYPE, userInfo.IdentifierID, datatypes.JSON(oauthLarkUserInfo))
		// Store the sso user info in redis for binding email
		s.Store.Set(ctx, fmt.Sprintf("BIND-EMAIL-%s-%s", idpName, userInfo.IdentifierID), studentID, store.BIND_EMAIL_EXP)
		systemSetting, err := s.Store.GetSystemSetting(ctx, store.WebsiteSettingType)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
		}

		webSetting, err := systemSetting.GetWebsiteSetting()
		if err != nil || webSetting == nil {
			return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
		}
		frontendURL := webSetting.FrontendURL
		// User email need to user input in frontend
		targetURL := fmt.Sprintf("%s/sso-bind-email?client_type=%s&idp_user_id=%s", frontendURL, idpName, userInfo.IdentifierID)

		// Redirect to frontend
		return c.Redirect(http.StatusTemporaryRedirect, targetURL)
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// BindEmailWithSSO bind email with SSO
//
// Bind email with SSO, it will check if the email is already registered, if not, create a new user with the email,
// if the email is already registered, we will bind the email with the user.
func (s *APIV1Service) BindEmailWithSSO(c echo.Context) error {
	ctx := c.Request().Context()

	clientType := c.QueryParam("client_type")
	idpUserID := c.QueryParam("idp_user_id")
	// Get email from form not query
	email := c.FormValue("email")
	redisKey := fmt.Sprintf("BIND-EMAIL-%s-%s", clientType, idpUserID)

	// Before bind email, check if the email is valid
	if !validator.ValidateEmail(email) {
		return echo.NewHTTPError(http.StatusBadRequest, response.UserEmailError)
	}

	idpUserInfo, err := s.Store.Get(ctx, redisKey)
	if err != nil || idpUserInfo == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}

	studentID := util.GetStudentIDFromEmail(email)
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}

	// User not found, Need to register to bind the sso id
	user, err := s.Store.UserByField("uid", studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}
	if user == nil {
		// TODO: Create user and profile,
		password, err := util.GenerateRandomString(20)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
		}
		if err := s.UserService.CreateUserAndProfile(email, password); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, response.CreateUserFail)
		}
	}

	// Bind email with SSO
	s.UpsetOauthInfo(studentID, clientType, idpUserID, datatypes.JSON(idpUserInfo))

	// Delete the redis key
	go s.Store.Delete(ctx, redisKey)

	// Set Token with expire time and return
	token, err := util.GenerateTokenWithExp(ctx, request.LoginJWTSubKey(studentID), s.Config.Secret, request.LOGIN_ACCESS_TOKEN_EXP)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	s.Store.Set(ctx, request.LoginTokenKey(studentID), token, request.LOGIN_ACCESS_TOKEN_EXP)
	response.SetCookie(c, request.AccessTokenCookieName, token)

	return c.JSON(http.StatusOK, response.Success(token))
}

// TODO: Implement this function for login and login with SSO
func (s *APIV1Service) doLogin(ctx context.Context, username, password string) error {
	return nil
}

func (s *APIV1Service) Register(c echo.Context) error {
	ctx := c.Request().Context()
	// Get Body from request
	password := c.FormValue("password")
	if password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	cookie, err := c.Cookie("REGISTER-TICKET")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}
	ticket := cookie.Value

	currentPhase, _ := s.Store.Get(ctx, ticket)
	if currentPhase != request.VERIFY_STATUS["VERIFY_ACCOUNT"] {
		return echo.NewHTTPError(http.StatusBadRequest, response.RegisterPhaseError)
	}

	username, usernameErr := util.IdentityFromToken(ticket, request.REGIST_TICKET_SUB, s.Config.Secret)
	if usernameErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}

	creErr := s.CreateUserAndProfile(username, password)
	if creErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}
	// set VERIFY_STATUS to 3 if successes
	s.Store.Set(ctx, ticket, request.VERIFY_STATUS["SUCCESS"], request.REGISTER_TICKET_EXP)
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) CheckVerifyCode(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.FormValue("verify_code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	// Get TICKET from cookies
	var ticket, flag string
	if cookie, err := c.Cookie("REGISTER-TICKET"); err == nil {
		ticket = cookie.Value
		flag = request.REGIST_TICKET_SUB
	} else if cookie, err := c.Cookie("RESETPWD-TICKET"); err == nil {
		ticket = cookie.Value
		flag = request.RESETPWD_TICKET_SUB
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}

	studentID := request.GetUsername(c.Request())
	codeError := s.UserService.CheckVerifyCode(ctx, ticket, code, flag, studentID)
	if codeError != nil {
		return echo.NewHTTPError(http.StatusBadRequest, codeError.Error())
	}
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) Verify(c echo.Context) error {
	ctx := c.Request().Context()
	// Username maybe email or studentID
	username := c.QueryParam("username")
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}
	// Capitalize the username
	username = strings.ToLower(username)

	flag, _ := strconv.Atoi(c.QueryParam("flag"))
	tKey := ""
	// 0 is register
	// 1 is login
	// 2 is resetPassword
	if flag == 0 {
		tKey = request.REGIST_TICKET_SUB
	} else if flag == 1 {
		tKey = request.LOGIN_TICKET_SUB
	} else if flag == 2 {
		tKey = request.RESETPWD_TICKET_SUB
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	ticket, err := s.UserService.VerifyAccount(ctx, username, flag)
	if err != nil {
		log.Errorf("Verify account fail: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	response.SetCookie(c, tKey, ticket)
	resMap := make(map[string]string)
	resMap[tKey] = ticket
	return c.JSON(http.StatusOK, response.Success(resMap))
}

func (s *APIV1Service) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	cookie, err := c.Cookie(request.LOGIN_TOKEN_SUB)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}
	token := cookie.Value
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}
	uid, err := util.IdentityFromToken(token, request.LOGIN_TOKEN_SUB, s.Config.Secret)
	if err != nil || uid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.InternalErr)
	}
	// Delete token from redis
	s.Store.Delete(ctx, request.LoginTokenKey(uid))
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) SendEmail(c echo.Context) error {
	ctx := c.Request().Context()
	// Get TICKET from cookies
	var ticket, flag string
	if cookie, err := c.Cookie("REGISTER-TICKET"); err == nil {
		ticket = cookie.Value
		flag = request.REGIST_TICKET_SUB
	} else if cookie, err := c.Cookie("RESETPWD-TICKET"); err == nil {
		ticket = cookie.Value
		flag = request.RESETPWD_TICKET_SUB
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, response.CheckTicketNotfound)
	}

	username, usernameErr := util.IdentityFromToken(ticket, flag, s.Config.Secret)
	// 错误处理机制写玉玉了
	// 我开始乱写了啊啊啊啊
	if usernameErr != nil {
		log.Errorf("username parse error: %s", usernameErr.Error())
		return echo.NewHTTPError(http.StatusUnauthorized, response.TicketNotCorrect)
	}
	// verify if the user email correct
	if !validator.ValidateEmail(username) {
		return echo.NewHTTPError(http.StatusBadRequest, response.UserEmailError)
	}

	var title string
	if flag == request.REGIST_TICKET_SUB {
		title = "确认电子邮件注册SAST-Link账户（无需回复）"
	} else {
		title = "确认电子邮件重置SAST-Link账户密码（无需回复）"
	}
	err := s.UserService.SendEmail(ctx, username, ticket, title)
	if err != nil {
		log.Errorf("Send email fail: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	} else {
		return c.JSON(http.StatusOK, response.Success(nil))
	}
}
