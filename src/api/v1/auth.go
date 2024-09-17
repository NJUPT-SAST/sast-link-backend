package v1

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
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

	// It will return [ErrNoCookie] if the cookie is not found
	cookie, err := c.Cookie(request.LOGIN_TICKET_SUB)
	if err != nil {
		return response.Error(c, response.TICKET_NOT_FOUND)
	}
	ticket := cookie.Value
	// Get username from ticket
	username, err := util.IdentityFromToken(ticket, request.LOGIN_TICKET_SUB)
	if err != nil {
		log.Errorf("Get username from ticket fail: %s", err.Error())
		return response.Error(c, response.TICKET_INVALID)
	}
	password := c.FormValue("password")
	if password == "" || username == "" {
		log.Errorf("Login fail, username: %s", username)
		return response.Error(c, response.REQUIRED_PARAMS)
	}

	uid, err := s.UserService.Login(username, password)
	if err != nil {
		log.Errorf("Login fail: %s", err.Error())
		return response.Error(c, response.LOGIN_FAILED)
	}

	// Set Token with expire time and return
	token, err := util.GenerateTokenWithExp(ctx, request.LoginJWTSubKey(uid), request.LOGIN_ACCESS_TOKEN_EXP)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}
	s.Store.Set(ctx, request.LoginTokenKey(uid), token, request.LOGIN_ACCESS_TOKEN_EXP)
	response.SetCookieWithExpire(c, request.AccessTokenCookieName, token, request.LOGIN_ACCESS_TOKEN_EXP)

	// Upset the token to database
	if err := s.Store.UpsetAccessTokensUserSetting(ctx, uid, token, ""); err != nil {
		log.Errorf("Failed to upset access token to database: %s", err.Error())
		return response.Error(c, response.INTENAL_ERROR)
	}

	return c.JSON(http.StatusOK, response.Success(map[string]string{request.AccessTokenCookieName: token}))
}

// LoginWithSSO login with SSO, it will exchange the token with the SSO identity provider and get the user info.
//
// If the user is not registered, it will redirect to the frontend to bind the email.
func (s *APIV1Service) LoginWithSSO(c echo.Context) error {
	ctx := c.Request().Context()
	// Get Idp name from query
	idpName := c.QueryParam("idp")
	identityProvider, err := s.Store.GetIdentityProviderByName(ctx, idpName)
	if err != nil || identityProvider == nil {
		return response.Error(c, "identity provider not found")
	}

	var userInfo *oauth2.IdentityProviderUserInfo
	if identityProvider.Type == oauth2.IDPTypeOAuth2 {
		oauth2Idp, err := oauth2.NewOauth2IdentityProvider(identityProvider)
		if err != nil {
			log.Errorf("Failed to create oauth2 identity provider: %s", err.Error())
			return response.Error(c, "failed to create oauth2 identity provider")
		}
		token, err := oauth2Idp.ExchangeToken(ctx, identityProvider.GetOauth2Setting(), c.QueryParam("redirect_url"), c.QueryParam("code"))
		if err != nil {
			return response.Error(c, "exchange token fail")
		}
		userInfo, err = oauth2Idp.UserInfo(ctx, identityProvider.GetOauth2Setting(), token)
		if err != nil {
			return response.Error(c, "get user info fail")
		}
	} else {
		// Now only support OAuth2
		return response.Error(c, fmt.Sprintf("identity provider type %s not supported", identityProvider.Type))
	}

	user, err := s.Store.OauthInfoByUID(idpName, userInfo.IdentifierID)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}
	// NOTE: Get user info from the sso identity provider.
	/* Here we have 4 cases:
		1. user didn't bind SSO, nor registered before.
	       then we save the user SSO info to redis,
	       and return to frontend bind email page
	       (that page will request a 'email' and 'valification_code',
	       request two api: `VerifyEmail` and `SendEmail`)
		2. user registered before, didn't bind SSO
		   same with above, also check in `BindEmailWithSSO`
		3. user registered and bound SSO before
		   this can be checked by query 2 times in this api:
		   one for oauth_info, then for user info by uid from oath_info.
		   This way return login token directly
	*/
	if user == nil {
		// Store the sso user info in redis for binding email
		s.Store.Set(ctx, fmt.Sprintf("BIND-EMAIL-%s-%s", idpName, userInfo.IdentifierID), userInfo.IdentifierID, store.BIND_EMAIL_EXP)
		systemSetting, err := s.Store.GetSystemSetting(ctx, config.WebsiteSettingType.String())
		if err != nil {
			log.Errorf("Get website setting fail: %s", err.Error())
			return response.Error(c, response.INTENAL_ERROR)
		}

		webSetting := systemSetting.GetWebsiteSetting()
		if webSetting == nil {
			return response.Error(c, response.INTENAL_ERROR)
		}

		response.SetCookieWithExpire(c, "idp", idpName, request.LOGIN_ACCESS_TOKEN_EXP)
		response.SetCookieWithExpire(c, "idp_user_id", userInfo.IdentifierID, request.LOGIN_ACCESS_TOKEN_EXP)

		return c.JSON(http.StatusOK, nil)
	} else { // bound SSO before
		uid := user.UserID
		token, err := util.GenerateTokenWithExp(c.Request().Context(), store.LoginJWTSubKey(uid), store.LOGIN_TOKEN_EXP)
		if err != nil {
			return response.Error(c, response.INTENAL_ERROR)
		}
		s.Store.Set(c.Request().Context(), store.LoginTokenKey(uid), token, store.LOGIN_TOKEN_EXP)
		return c.JSON(http.StatusOK, map[string]string{store.LOGIN_TOKEN_SUB: token})
	}
}

// BindEmailWithSSO bind email with SSO
//
// Bind email with SSO, it will check if the email is already registered, if not, create a new user with the email,
// if the email is already registered, we will bind the email with the user.
func (s *APIV1Service) BindEmailWithSSO(c echo.Context) error {
	ctx := c.Request().Context()

	clientTypeCookie, err1 := c.Cookie("idp")
	idpUserIDCookie, err2 := c.Cookie("idp_user_id")
	ticketCookie, err3 := c.Cookie(request.OAUTH_CHECK_EMAIL_SUB)

	email := c.FormValue("email")

	if err1 != nil || err2 != nil || err3 != nil || email == "" || idpUserIDCookie.Value == "" || clientTypeCookie.Value == "" {
		return response.Error(c, response.REQUIRED_PARAMS)
	}

	ticket := ticketCookie.Value
	clientType := clientTypeCookie.Value
	idpUserID := idpUserIDCookie.Value

	currentPhase, err := s.Store.Get(ctx, ticket)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}
	if currentPhase != request.VERIFY_STATUS["VERIFY_CAPTCHA"] {
		return response.Error(c, "phase error, please verify email first!")
	}

	ticketEmail, err := util.IdentityFromToken(ticket, request.OAUTH_CHECK_EMAIL_SUB)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}

	if !validator.ValidateEmail(email) || email != ticketEmail {
		return response.Error(c, response.EMAIL_INVALID)
	}

	redisKey := fmt.Sprintf("BIND-EMAIL-%s-%s", clientType, idpUserID)
	idpUserInfo, err := s.Store.Get(ctx, redisKey)
	if err != nil || idpUserInfo == "" {
		return response.Error(c, "bind email redis key not exist")
	}

	studentID := util.GetStudentIDFromEmail(email)
	if studentID == "" {
		return response.Error(c, response.EMAIL_INVALID)
	}

	user, err := s.Store.UserByField(ctx, "uid", studentID)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}
	// user not registered
	// we generate a new account, and bind it to user email
	if user == nil {
		password, err := util.GenerateRandomString(20)
		if err != nil {
			return response.Error(c, response.INTENAL_ERROR)
		}
		if err := s.UserService.CreateUserAndProfile(email, password); err != nil {
			return response.Error(c, "create user fail")
		}
		// Bind email with SSO
		s.UpsetOauthInfo(studentID, clientType, idpUserID, datatypes.JSON(idpUserInfo))
	} else {
		// user registered but not bound before
		// we bind his/her SSO info to his/her email
		s.UpsetOauthInfo(studentID, clientType, idpUserID, datatypes.JSON(idpUserInfo))
	}

	// Delete the redis key
	go s.Store.Delete(ctx, redisKey)

	// Set Token with expire time and return
	token, err := util.GenerateTokenWithExp(ctx, request.LoginJWTSubKey(studentID), request.LOGIN_ACCESS_TOKEN_EXP)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}
	s.Store.Set(ctx, request.LoginTokenKey(studentID), token, request.LOGIN_ACCESS_TOKEN_EXP)
	response.SetCookieWithExpire(c, request.AccessTokenCookieName, token, request.LOGIN_ACCESS_TOKEN_EXP)

	return c.JSON(http.StatusOK, response.Success(map[string]string{request.AccessTokenCookieName: token}))
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
		return response.Error(c, response.REQUIRED_PARAMS)
	}

	cookie, err := c.Cookie(request.REGIST_TICKET_SUB)
	if err != nil {
		return response.Error(c, response.TICKET_NOT_FOUND)
	}
	ticket := cookie.Value

	currentPhase, err := s.Store.Get(ctx, ticket)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}
	switch currentPhase {
	case request.VERIFY_STATUS["VERIFY_ACCOUNT"], request.VERIFY_STATUS["SEND_EMAIL"]:
		return response.Error(c, "please check your email to verify your account first")
	case request.VERIFY_STATUS["SUCCESS"]:
		return response.Error(c, response.USER_EXIST)
	case "":
		return response.Error(c, response.TICKET_NOT_FOUND)
	}

	studentID, err := util.IdentityFromToken(ticket, request.REGIST_TICKET_SUB)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}

	if err := s.CreateUserAndProfile(studentID, password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create user fail")
	}
	// set VERIFY_STATUS to 3 if successes
	s.Store.Set(ctx, ticket, request.VERIFY_STATUS["SUCCESS"], request.REGISTER_TICKET_EXP)
	log.Debugf("User [%s] register success", studentID)
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) CheckVerifyCode(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.FormValue("verify_code")
	if code == "" {
		return response.Error(c, response.REQUIRED_PARAMS)
	}

	// Get TICKET from cookies
	var ticket, flag string
	if cookie, err := c.Cookie(request.REGIST_TICKET_SUB); err == nil {
		ticket = cookie.Value
		flag = request.REGIST_TICKET_SUB
	} else if cookie, err := c.Cookie(request.RESETPWD_TICKET_SUB); err == nil {
		ticket = cookie.Value
		flag = request.RESETPWD_TICKET_SUB
	} else if cookie, err := c.Cookie(request.OAUTH_CHECK_EMAIL_SUB); err == nil {
		ticket = cookie.Value
		flag = request.OAUTH_CHECK_EMAIL_SUB
	} else {
		return response.Error(c, response.TICKET_NOT_FOUND)
	}

	studentID, err := util.IdentityFromToken(ticket, flag)
	if err != nil {
		return response.Error(c, response.INTENAL_ERROR)
	}

	status, err := s.Store.Get(ctx, ticket)
	if err != nil || status == "" {
		return response.Error(c, "failed to get current status")
	}

	if err := s.UserService.CheckVerifyCode(ctx, status, code, flag, studentID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.VERIFY_CODE_INCORRECT)
	}

	// Update the status of the ticket
	s.Store.Set(ctx, ticket, request.VERIFY_STATUS["VERIFY_CAPTCHA"], request.REGISTER_TICKET_EXP)
	log.Debugf("User [%s] check verify code success", studentID)
	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) Verify(c echo.Context) error {
	ctx := c.Request().Context()
	// Username maybe email or studentID
	username := c.QueryParam("username")
	if username == "" {
		return response.Error(c, response.REQUIRED_PARAMS)
	}
	// Capitalize the username
	username = strings.ToLower(username)

	log.Infof("Verify username: %s", username)

	flag, _ := strconv.Atoi(c.QueryParam("flag"))
	tKey := ""
	var tExp time.Duration
	// 0 is register
	// 1 is login
	// 2 is resetPassword
	if flag == 0 {
		tKey = request.REGIST_TICKET_SUB
		tExp = request.REGISTER_TICKET_EXP
	} else if flag == 1 {
		tKey = request.LOGIN_TICKET_SUB
		tExp = request.LOGIN_TICKET_EXP
	} else if flag == 2 {
		tKey = request.RESETPWD_TICKET_SUB
		tExp = request.RESETPWD_TICKET_EXP
	} else {
		return response.Error(c, response.REQUIRED_PARAMS)
	}

	ticket, err := s.UserService.VerifyAccount(ctx, username, flag)
	if err != nil {
		log.Errorf("Verify account fail: %s", err.Error())
		return response.Error(c, "verify account fail")
	}

	response.SetCookieWithExpire(c, tKey, ticket, tExp)
	return c.JSON(http.StatusOK, response.Success(map[string]string{tKey: ticket}))
}

// VerifyEmail Verify if this email is valid,
// then set oauth check email ticket in cookie.
//
// mainly used for BindEmailWithSSO.
func (s *APIV1Service) VerifyEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if !validator.ValidateEmail(email) {
		return response.Error(c, response.EMAIL_INVALID)
	}

	ticket, err := util.GenerateTokenWithExp(c.Request().Context(), request.BindSSOSubKey(util.GetStudentIDFromEmail(email)), request.LOGIN_TICKET_EXP)
	if err != nil {
		log.Errorf("Verify email fail: %s", err.Error())
		return response.Error(c, "verify email fail")
	}

	s.Store.Set(c.Request().Context(), ticket, request.VERIFY_STATUS["VERIFY_ACCOUNT"], request.VERIFY_CODE_EXP)
	response.SetCookieWithExpire(c, request.OAUTH_CHECK_EMAIL_SUB, ticket, request.LOGIN_TICKET_EXP)

	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	accessToken := request.GetAccessToken(c.Request())
	if accessToken == "" {
		return response.Error(c, response.UNAUTHORIZED)
	}
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return response.Error(c, response.UNAUTHORIZED)
	}

	// Delete the access token
	if err := s.UserService.DeleteUserAccessToken(ctx, studentID, accessToken); err != nil {
		log.Errorf("Delete access token fail: %s", err.Error())
	}

	response.SetCookieWithExpire(c, request.AccessTokenCookieName, "", 0)

	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *APIV1Service) SendEmail(c echo.Context) error {
	ctx := c.Request().Context()
	// Get TICKET from cookies
	var ticket, flag string
	if cookie, err := c.Cookie(request.REGIST_TICKET_SUB); err == nil {
		ticket = cookie.Value
		flag = request.REGIST_TICKET_SUB
	} else if cookie, err := c.Cookie(request.RESETPWD_TICKET_SUB); err == nil {
		ticket = cookie.Value
		flag = request.RESETPWD_TICKET_SUB
	} else if cookie, err := c.Cookie(request.OAUTH_CHECK_EMAIL_SUB); err == nil {
		ticket = cookie.Value
		flag = request.OAUTH_CHECK_EMAIL_SUB
	} else {
		return response.Error(c, response.TICKET_NOT_FOUND)
	}
	log.Debugf("SendEmail::[ticket: %s] [flag: %s]", ticket, flag)

	studentID, err := util.IdentityFromToken(ticket, flag)
	// 错误处理机制写玉玉了
	// 我开始乱写了啊啊啊啊
	if err != nil {
		log.Errorf("username parse error: %s", err.Error())
		return response.Error(c, response.INTENAL_ERROR)
	}

	// Verify if the user email correct
	log.Debugf("SendEmail::[studentID: %s]", studentID)
	if !validator.ValidateStudentID(studentID) {
		log.Errorf("student not valid : %s", studentID)
		return echo.NewHTTPError(http.StatusBadRequest, response.EMAIL_INVALID)
	}

	var title string
	if flag == request.REGIST_TICKET_SUB {
		title = "确认电子邮件注册SAST-Link账户（无需回复）"
	} else if flag == request.OAUTH_CHECK_EMAIL_SUB {
		title = "确认电子邮件绑定SSO账号（无需回复）"
	} else {
		title = "确认电子邮件重置SAST-Link账户密码（无需回复）"
	}
	status, err := s.Store.Get(ctx, ticket)
	if err != nil || status == "" {
		log.Errorf("should not in current phase: %v", err)
		return response.Error(c, response.INTENAL_ERROR)
	}

	email := util.UserNameToEmail(studentID)
	if err := s.UserService.SendEmail(ctx, email, status, title); err != nil {
		log.Errorf("Send email fail: %s", err.Error())
		return response.Error(c, err)
	}

	// Update the status of the ticket
	s.Store.Set(ctx, ticket, request.VERIFY_STATUS["SEND_EMAIL"], request.REGISTER_TICKET_EXP)
	log.Debugf("send email to [%s] successfully", email)
	return c.JSON(http.StatusOK, response.Success(nil))
}
