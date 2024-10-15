package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/util"
)

type AuthInterceptor struct {
	store  *store.Store
	secret string

	log *zap.Logger
}

func NewAuthInterceptor(store *store.Store, secret string) *AuthInterceptor {
	logger := log.NewLogger(log.WithModule("auth"), log.WithLayer("middleware"))
	return &AuthInterceptor{store: store, secret: secret, log: logger}
}

func (m *AuthInterceptor) AuthenticationInterceptor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r, _ := c.Request(), c.Response().Writer
		ctx := r.Context()

		defer func() {
			m.log.Debug("Incomming request", zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.String("client_ip", c.RealIP()),
				zap.String("user_agent", r.UserAgent()))
		}()

		if r.URL == nil {
			return response.Error(c, response.InternalError)
		}
		if isUnauthorizeAllowed(r.URL.Path) {
			return next(c)
		}
		clientIP := c.RealIP()
		accesstoken := getAccessToken(r)

		username, err := m.authenticate(r.Context(), accesstoken)
		if err != nil {
			m.log.Error("Failed to authenticate",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.Error(err))
			return response.Error(c, response.UNAUTHORIZED)
		}
		user, err := m.store.UserInfo(ctx, username)
		if err != nil {
			m.log.Error("Failed to get user",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.Error(err))
			return response.Error(c, response.UNAUTHORIZED)
		}

		if user == nil {
			m.log.Error("User not found",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.Error(err))
			return response.Error(c, response.UNAUTHORIZED)
		}

		m.log.Debug("User entered",
			zap.String("client_ip", clientIP),
			zap.String("user_agent", r.UserAgent()),
			zap.String("username", username),
		)

		// if isOnlyForAdminAllowedPath(r.URL.Path) && user.Role != model.RoleHost && user.Role != model.RoleAdmin {
		// 	log.DebugWithFields("User is not allowed to access this path",
		// 		log.Fields{"client_ip": clientIP, "user_agent": r.UserAgent(), "username": username, "role": user.Role.String()})
		// 	return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("User is not allowed to access this path"))
		// }

		// m.store.SetLastLogin(user.ID)
		// m.store.SetAPIKeyUsedTimeStamp(user.ID, accesstoken)

		// c.Set()
		// Must use string to store in context
		ctx = context.WithValue(ctx, request.UserIDContextKey, strconv.Itoa(int(user.ID)))
		// Here we use the student ID as the username
		ctx = context.WithValue(ctx, request.UserNameContextKey, *user.UID) // c.Set(string(request.UserNameContextKey), *user.Uid)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
		// Set the access token in the context for oauth server use
		ctx = context.WithValue(ctx, request.AccessTokenContextKey, accesstoken)
		// ctx = context.WithValue(ctx, request.UserRolesContextKey, user.Role.String())

		c.SetRequest(r.WithContext(ctx))

		return next(c)
	}
}

func (m *AuthInterceptor) authenticate(ctx context.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", errors.New("no access token provided")
	}

	// Validate the login access token
	userID, err := util.IdentityFromToken(accessToken, request.AccessTokenCookieName)
	if err != nil {
		return "", errors.Wrap(err, "invalid or expired access token")
	}
	user, err := m.store.UserInfo(ctx, userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user")
	}
	if user == nil {
		return "", errors.Errorf("user not found with ID: %s", userID)
	}

	accessTokens, err := m.store.GetUserAccessTokens(ctx, userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user access tokens")
	}

	if !validateAccessToken(accessToken, accessTokens) {
		return "", errors.New("invalid access token")
	}

	return *user.UID, nil
}

// getAccessToken will get the access token from the request, it will first check the Authorization header,
// then check the cookie header.
func getAccessToken(r *http.Request) string {
	// Check the HTTP Authorization header first
	authorizationHeaders := r.Header.Get("Authorization")
	// Check bearer token
	if authorizationHeaders != "" {
		splitToken := strings.Split(authorizationHeaders, "Bearer ")
		if len(splitToken) == 2 {
			return splitToken[1]
		}
	}

	// Check the cookie header
	var accessToken string
	for cookie := range r.Cookies() {
		if r.Cookies()[cookie].Name == request.AccessTokenCookieName {
			accessToken = r.Cookies()[cookie].Value
		}
	}
	return accessToken
}

func validateAccessToken(accessTokenString string, userAccessTokens []*store.UserSettingAccessToken) bool {
	for _, userAccessToken := range userAccessTokens {
		if accessTokenString == userAccessToken.AccessToken {
			return true
		}
	}
	return false
}
