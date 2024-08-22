package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
)

type OAuthServer struct {
	Srv         *server.Server
	TokenStore  *pg.TokenStore
	ClientStore *pg.ClientStore
}

func NewOAuthServer(ctx context.Context, config *config.Config) (*OAuthServer, error) {
	// postgresql://username:password@ip:port/dbname
	dsn := fmt.Sprintf(`postgres://%s:%s@%s:%d/%s?sslmode=disable`,
		config.PostgresUser,
		config.PostgresPWD,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDB,
	)
	pgxConn, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Panicf("Failed to connect to PostgreSQL for OAuth server: %s", err)
		return nil, err
	}

	tokenAdapter := pgx4adapter.NewPool(pgxConn)
	tokenStore, err := pg.NewTokenStore(tokenAdapter, pg.WithTokenStoreGCInterval(time.Minute))
	if err != nil {
		log.Panicf("Failed to create token store for OAuth server: %s", err)
		return nil, err
	}
	clientAdapter := pgx4adapter.NewPool(pgxConn)
	clientStore, err := pg.NewClientStore(clientAdapter)
	if err != nil {
		log.Panicf("Failed to create client store for OAuth server: %s", err)
		return nil, err
	}

	mg := manage.NewDefaultManager()
	mg.MapTokenStorage(tokenStore)
	mg.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	// use PostgreSQL client store with pgx.Connection adapter
	mg.MapClientStorage(clientStore)

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
	s.Srv.SetInternalErrorHandler(InternalErrorHandler)
	s.Srv.SetResponseErrorHandler(ResponseErrorHandler)
	s.Srv.SetResponseTokenHandler(ResponseTokenHandler)
}

func InternalErrorHandler(err error) (re *errors.Response) {
	log.Errorf("Oauth2 Server ::: InternalErrorHandler:[%s]", err.Error())
	error := errors.NewResponse(err, http.StatusInternalServerError)
	error.ErrorCode = 500
	error.StatusCode = http.StatusInternalServerError
	error.Description = err.Error()
	return error
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
	if data["error"] != nil {
		var errMsg string
		var errCode int
		if data["error_description"] != nil {
			errMsg = data["error_description"].(string)
		}
		if data["error_code"] != nil {
			errCode = data["error_code"].(int)
		}
		err := response.LocalError{
			ErrCode: errCode,
			ErrMsg:  errMsg,
		}
		log.Errorf("Oauth2 ::: ResponseTokenHandler:error:[%s]", err)
		return json.NewEncoder(w).Encode(response.Failed(err))
	} else {
		return json.NewEncoder(w).Encode(response.Success(data))
	}
}

// ClientStoreItem data item
type ClientStoreItem struct {
	ID     string `db:"id"`
	Secret string `db:"secret"`
	Domain string `db:"domain"`
	Data   []byte `db:"data"`
}

// CreateClient creates a new client
func (s *APIV1Service) CreateClient(c echo.Context) error {
	redirectURI := c.FormValue("redirect_uri")
	if redirectURI == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	// token := c.GetHeader("TOKEN")
	studentID := request.GetUsername(c.Request())
	if studentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, response.RequestParamError)
	}

	clientID := util.GenerateUUID()
	secret, err := util.GenerateRandomString(32)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	if s.OAuthServer.ClientStore.Create(&models.Client{
		ID:     clientID,
		Secret: secret,
		Domain: redirectURI,
		UserID: studentID,
	}) != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	return c.JSON(http.StatusOK, response.Success(map[string]string{
		"client_id":     clientID,
		"client_secret": secret,
	}))
}

// OauthUserInfo response user info
func (s *APIV1Service) OauthUserInfo(c echo.Context) error {
	ctx := c.Request().Context()
	if request.GetIsAuthenticated(c.Request()) == false {
		return c.JSON(http.StatusOK, response.Failed(response.Unauthorized))
	}
	accessToken := request.GetAccessToken(c.Request())
	if accessToken == "" {
		return c.JSON(http.StatusOK, response.Failed(response.Unauthorized))
	}
	mg := s.OAuthServer.Srv.Manager
	ti, err := mg.LoadAccessToken(ctx, accessToken)
	if err != nil {
		return c.JSON(http.StatusOK, response.Failed(response.Unauthorized))
	}
	// TODO: scope check
	ti.GetScope()

	user, err := s.OauthService.OauthUserInfo(ti.GetUserID())
	if err != nil {
		log.Errorf("Failed to get user info: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	profileInfo, serErr := s.ProfileService.GetProfileInfo(*user.Uid)
	if serErr != nil {
		log.Errorf("Failed to get profile info: %s", serErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	if dep, org, getOrgErr := s.ProfileService.GetProfileOrg(profileInfo.OrgId); getOrgErr != nil {
		log.Errorf("Failed to get profile org: %s", getOrgErr)
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	} else {
		return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
			"nickname": profileInfo.Nickname,
			"userId":   user.Uid,
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
}

// Authorize will redirect user to login page if user not login
func (s *APIV1Service) Authorize(c echo.Context) error {
	r := c.Request()
	w := c.Response().Writer
	_ = r.ParseForm()
	// Redirect user to login page if user not login or
	// get code directly if user has logged in
	if s.OAuthServer.Srv.HandleAuthorizeRequest(w, r) != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// AccessToken returns access token
func (s *APIV1Service) AccessToken(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	if s.OAuthServer.Srv.HandleTokenRequest(w, r) != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

// RefreshToken returns new access token
func (s *APIV1Service) RefreshToken(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	err := s.OAuthServer.Srv.HandleTokenRequest(w, r)
	if err == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.InternalErr)
	}

	return c.JSON(http.StatusOK, response.Success(nil))
}

func (s *OAuthServer) clientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
	_ = r.ParseMultipartForm(0)
	_ = r.ParseForm()
	if r.Form.Get("grant_type") == "refresh_token" {
		ti, err := s.Srv.Manager.LoadRefreshToken(r.Context(), r.Form.Get("refresh_token"))
		if err != nil {
			return "", "", response.RefreshTokenErr
		}
		clientID = ti.GetClientID()
		if clientID == "" {
			return "", "", response.ClientErr
		}
		cli, err := s.Srv.Manager.GetClient(r.Context(), clientID)
		if err != nil {
			return "", "", response.ClientErr
		}
		clientSecret = cli.GetSecret()
		if clientSecret == "" {
			return "", "", response.ClientErr
		}
		return clientID, clientSecret, nil
	}
	clientID = r.Form.Get("client_id")
	if clientID == "" {
		return "", "", response.ClientErr
	}
	clientSecret = r.Form.Get("client_secret")
	if clientSecret == "" {
		return "", "", response.ClientErr
	}
	return clientID, clientSecret, nil
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	if !request.GetIsAuthenticated(r) {
		return "", response.Unauthorized
	}

	stuID := request.GetUsername(r)
	if stuID == "" {
		return "", response.UsernameError
	}

	return stuID, nil
}
