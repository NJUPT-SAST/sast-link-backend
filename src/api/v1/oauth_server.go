package v1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/labstack/echo/v4"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/util"
)

type OAuthServer struct {
	Srv         *server.Server
	TokenStore  *store.TokenStore
	ClientStore *store.ClientStore
}

func NewOAuthServer(_ context.Context, dbStore store.Store) (*OAuthServer, error) {
	tokenStore := store.NewTokenStore(dbStore, store.WithTokenStoreGCInterval(5*time.Minute))
	clientStore := store.NewClientStore(dbStore)

	mg := manage.NewDefaultManager()
	mg.MapTokenStorage(tokenStore)
	mg.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	mg.MapClientStorage(clientStore)
	// Custom URI validation handler, we need to use multiple redirect uri
	mg.SetValidateURIHandler(dbStore.ValidateURIHandler)

	srv := server.NewServer(server.NewConfig(), mg)

	return &OAuthServer{
		Srv:         srv,
		TokenStore:  tokenStore,
		ClientStore: clientStore,
	}, nil
}

func (s *OAuthServer) SetHandler() {
	s.Srv.SetClientInfoHandler(s.clientInfoHandler)
	s.Srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	// according to spec, servers should respond status 400 in error case:
	// RFC 6749 https://www.rfc-editor.org/rfc/rfc6749#section-5.2
	// TODO: set error response status code to 400
	// s.Srv.SetInternalErrorHandler(InternalErrorHandler)
	// s.Srv.SetResponseErrorHandler(ResponseErrorHandler)

	// RFC 6749 provides a example of successful response:
	// RFC 6749 https://www.rfc-editor.org/rfc/rfc6749#section-5.1
	// s.Srv.SetResponseTokenHandler(ResponseTokenHandler)

	s.Srv.SetPreRedirectErrorHandler(PreRedirectErrorHandler)
}

// nolint
func PreRedirectErrorHandler(w http.ResponseWriter, req *server.AuthorizeRequest, err error) error {
	log.Errorf("Oauth2 Server ::: PreRedirectErrorHandler:[%s]", err.Error())
	return err
}

func InternalErrorHandler(err error) (re *errors.Response) {
	log.Errorf("Oauth2 Server ::: InternalErrorHandler:[%s]", err.Error())
	reErr := errors.NewResponse(err, http.StatusInternalServerError)
	reErr.ErrorCode = 500
	reErr.StatusCode = http.StatusInternalServerError
	reErr.Description = err.Error()
	return reErr
}

func ResponseErrorHandler(re *errors.Response) {
	log.Errorf("Oauth2 ::: ResponseErrorHandler:[%s]", re.Error.Error())
}

func ResponseTokenHandler(w http.ResponseWriter, data map[string]interface{}, header http.Header, statusCode ...int) error {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	for key := range header {
		w.Header().Set(key, header.Get(key))
	}

	status := http.StatusOK
	if len(statusCode) > 0 && statusCode[0] > 0 {
		status = statusCode[0]
	}

	w.WriteHeader(status)

	// If there is an error, return error message
	if data["error"] != nil {
		var errMsg string
		var errCode int

		if desc, ok := data["error_description"].(string); ok {
			errMsg = desc
		}

		if code, ok := data["error_code"].(int); ok {
			errCode = code
		}

		err := response.LocalError{
			ErrCode: errCode,
			ErrMsg:  errMsg,
		}
		log.Errorf("Oauth2 ::: ResponseTokenHandler:error:[%s]", err)
		return json.NewEncoder(w).Encode(response.Failed(err))
	}

	if token, ok := data["access_token"].(string); ok {
		return json.NewEncoder(w).Encode(map[string]string{"access_token": token})
	}

	err := response.LocalError{
		ErrCode: 400,
		ErrMsg:  "access_token not found",
	}

	return json.NewEncoder(w).Encode(response.Failed(err))
}

// OauthUserInfo response user info.
func (s *APIV1Service) OauthUserInfo(c echo.Context) error {
	ctx := c.Request().Context()
	authorization := c.Request().Header.Get("Authorization")
	// Remove "Bearer " prefix
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if accessToken == "" {
		return c.JSON(http.StatusUnauthorized, response.Failed(response.UNAUTHORIZED))
	}
	mg := s.OAuthServer.Srv.Manager
	ti, err := mg.LoadAccessToken(ctx, accessToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Failed(response.UNAUTHORIZED))
	}
	// TODO: scope check
	ti.GetScope()

	user, err := s.OauthService.OauthUserInfo(ctx, ti.GetUserID())
	if err != nil {
		log.Errorf("Failed to get user info: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, response.Failed(response.UNAUTHORIZED))
	}

	profileInfo, err := s.ProfileService.GetProfileInfo(*user.UID)
	if err != nil {
		log.Errorf("Failed to get profile info: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	dep, org, err := s.ProfileService.GetProfileOrg(profileInfo.OrgID)
	if err != nil {
		log.Errorf("Failed to get profile org: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"nickname": profileInfo.Nickname,
		"userId":   user.UID,
		"dep":      dep,
		"org":      org,
		"email":    profileInfo.Email,
		"avatar":   profileInfo.Avatar,
		"bio":      profileInfo.Bio,
		"link":     profileInfo.Link,
		"badge":    profileInfo.Badge,
		"hide":     profileInfo.Hide,
	}))
}

// Authorize will redirect user to login page if user not login.
func (s *APIV1Service) Authorize(c echo.Context) error {
	r := c.Request()
	w := c.Response().Writer
	_ = r.ParseForm()

	// Redirect user to login page if user not login or
	// get code directly if user has logged in
	return s.OAuthServer.Srv.HandleAuthorizeRequest(w, r)
}

// AccessToken returns access token.
func (s *APIV1Service) AccessToken(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()

	return s.OAuthServer.Srv.HandleTokenRequest(w, r)
}

// RefreshToken returns new access token.
func (s *APIV1Service) RefreshToken(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	return s.OAuthServer.Srv.HandleTokenRequest(w, r)
}

// CreateClient creates a new client.
func (s *APIV1Service) CreateClient(c echo.Context) error {
	ctx := c.Request().Context()

	redirectURI := c.FormValue("redirect_uri")
	if redirectURI == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	uid := request.GetUserID(c.Request())
	if uid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	clientID := util.GenerateUUID()
	secret, err := util.GenerateRandomString(32)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.RequiredParams))
	}

	clientName := c.FormValue("client_name")
	if clientName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	// clientDesc is optional
	clientDesc := c.FormValue("client_desc")

	if s.OAuthServer.ClientStore.Create(ctx, &models.Client{
		ID:     clientID,
		Secret: secret,
		Domain: redirectURI,
		UserID: uid,
	}, clientName, clientDesc) != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.RequiredParams))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]string{
		"client_id":     clientID,
		"client_secret": secret,
	}))
}

// AddRedirectURI add redirect uri to client.
func (s *APIV1Service) AddRedirectURI(c echo.Context) error {
	ctx := c.Request().Context()

	uid := request.GetUserID(c.Request())

	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	if clientID == "" || redirectURI == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	if err := s.OAuthServer.ClientStore.AddRedirectURI(ctx, uid, clientID, redirectURI); err != nil {
		log.Errorf("Failed to add redirect uri: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// ListClient returns cilent list.
func (s *APIV1Service) ListClient(c echo.Context) error {
	ctx := c.Request().Context()

	uid := request.GetUserID(c.Request())

	find := store.FindClientRequest{
		UserID: uid,
	}
	list, err := s.OAuthServer.ClientStore.ListClient(ctx, find)
	if err != nil {
		log.Errorf("Failed to list client: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	return c.JSON(http.StatusOK, response.Success(list))
}

func (s *APIV1Service) GetClient(c echo.Context) error {
	ctx := c.Request().Context()

	clientID := c.QueryParam("client_id")
	if clientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	uid := request.GetUserID(c.Request())
	if !s.OAuthServer.ClientStore.CheckClientOwner(ctx, clientID, uid) {
		return echo.NewHTTPError(http.StatusForbidden, response.Failed(response.FORBIDDEN))
	}

	request := store.FindClientRequest{
		ID:     clientID,
		UserID: uid,
	}

	client, err := s.OAuthServer.ClientStore.GetClient(ctx, request)
	if err != nil {
		log.Errorf("Failed to get client: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	return c.JSON(http.StatusOK, response.Success(client))
}

// UpdateClient updates client info.
func (s *APIV1Service) UpdateClient(c echo.Context) error {
	ctx := c.Request().Context()

	clientID := c.FormValue("client_id")
	clientName := c.FormValue("client_name")
	clientDesc := c.FormValue("client_desc")
	redirectURIStr := c.FormValue("redirect_uri")
	if clientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	uid := request.GetUserID(c.Request())

	if !s.OAuthServer.ClientStore.CheckClientOwner(ctx, clientID, uid) {
		return echo.NewHTTPError(http.StatusForbidden, response.Failed(response.FORBIDDEN))
	}

	redirectURIs := strings.Split(redirectURIStr, ",")
	for uir := range redirectURIs {
		if !util.CheckRedirectURI(redirectURIs[uir]) {
			return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.InvalidRedirectURI))
		}
	}

	requestClient := store.UpdateClientRequest{
		ID:           clientID,
		Name:         clientName,
		Desc:         clientDesc,
		RedirectURIs: redirectURIs,
		UserID:       uid,
	}

	if err := s.OAuthServer.ClientStore.UpdateClient(ctx, requestClient); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// DelClient deletes client.
func (s *APIV1Service) DelClient(c echo.Context) error {
	ctx := c.Request().Context()

	clientID := c.FormValue("client_id")
	if clientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.Failed(response.RequiredParams))
	}

	uid := request.GetUserID(c.Request())

	// Check whether the client belongs to the user
	if !s.OAuthServer.ClientStore.CheckClientOwner(ctx, clientID, uid) {
		return echo.NewHTTPError(http.StatusForbidden, response.Failed(response.FORBIDDEN))
	}

	if err := s.OAuthServer.ClientStore.DeleteClient(ctx, clientID, uid); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.Failed(response.InternalError))
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// clientInfoHandler returns client id and client secret.
func (s *OAuthServer) clientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
	if r.Form.Get("grant_type") == "refresh_token" {
		ti, err := s.Srv.Manager.LoadRefreshToken(r.Context(), r.Form.Get("refresh_token"))
		// Here errors package is from go-oauth2
		if err != nil {
			return "", "", errors.New("refresh token not found")
		}
		clientID = ti.GetClientID()
		if clientID == "" {
			return "", "", errors.New("client id not found")
		}
		cli, err := s.Srv.Manager.GetClient(r.Context(), clientID)
		if err != nil {
			return "", "", errors.New("client not found")
		}
		clientSecret = cli.GetSecret()
		if clientSecret == "" {
			return "", "", errors.New("client secret not found")
		}
		return clientID, clientSecret, nil
	}
	clientID, clientSecret, ok := parseBasicAuth(r.Header.Get("Authorization"))
	if !ok {
		return "", "", errors.New("client id or client secret not found")
	}

	log.Debugf("Oauth2 Server ::: client_id:[%s]", clientID)

	if clientID == "" {
		return "", "", errors.New("client id not found")
	}
	if clientSecret == "" {
		return "", "", errors.New("client secret not found")
	}
	return clientID, clientSecret, nil
}

// nolint
// userAuthorizeHandler get user id from request.
func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	if !request.GetIsAuthenticated(r) {
		return "", response.UNAUTHORIZED
	}

	stuID := request.GetUsername(r)
	if stuID == "" {
		return "", response.UserNotFound
	}

	return stuID, nil
}

// See 2 of the HTTP Authentication RFC 2617: https://www.rfc-editor.org/rfc/rfc2617
func parseBasicAuth(authHeader string) (username, password string, ok bool) {
	// Remove "Basic " prefix
	const prefix = "Basic "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", "", false
	}
	encodedCredentials := strings.TrimPrefix(authHeader, prefix)

	// Base64 decode
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", "", false
	}

	// Split username and password
	credentials := string(decoded)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}
