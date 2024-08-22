package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type AuthInterceptor struct {
	store  *store.Store
	secret string
}

func NewAuthInterceptor(store *store.Store, secret string) *AuthInterceptor {
	return &AuthInterceptor{store: store, secret: secret}
}

func (m *AuthInterceptor) AuthenticationInterceptor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r, _ := c.Request(), c.Response().Writer
		if isUnauthorizeAllowed(r.RequestURI) {
			return next(c)
		}
		clientIP := c.RealIP()
		accesstoken := getAccessToken(r)

		username, err := m.authenticate(r.Context(), accesstoken)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("Access unauthorized"))
		}
		user, err := m.store.UserInfo(username)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Failed to get user"))
		}

		if user == nil {
			log.DebugWithFields("User not found",
				log.Fields{"client_ip": clientIP, "user_agent": r.UserAgent(), "username": username})
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("User not found"))
		}

		log.DebugWithFields("User found",
			log.Fields{"client_ip": clientIP, "user_agent": r.UserAgent(), "username": username})

		// if isOnlyForAdminAllowedPath(r.URL.Path) && user.Role != model.RoleHost && user.Role != model.RoleAdmin {
		// 	log.DebugWithFields("User is not allowed to access this path",
		// 		log.Fields{"client_ip": clientIP, "user_agent": r.UserAgent(), "username": username, "role": user.Role.String()})
		// 	return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("User is not allowed to access this path"))
		// }

		// m.store.SetLastLogin(user.ID)
		// m.store.SetAPIKeyUsedTimeStamp(user.ID, accesstoken)

		// Set user context
		ctx := r.Context()

		// c.Set()
		// Must use string to store in context
		ctx = context.WithValue(ctx, request.UserIDContextKey, strconv.Itoa(int(user.ID)))
		// Here we use the student ID as the username
		ctx = context.WithValue(ctx, request.UserNameContextKey, user.Uid)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
		// Set the access token in the context for oauth server use
		ctx = context.WithValue(ctx, request.AccessTokenContextKey, accesstoken)
		// ctx = context.WithValue(ctx, request.UserRolesContextKey, user.Role.String())

		return next(c)
	}
}

func (m *AuthInterceptor) authenticate(ctx context.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", errors.New("no access token provided")
	}

	// Validate the login access token
	userID, err := util.IdentityFromToken(accessToken, request.AccessTokenCookieName, m.secret)
	if err != nil {
		return "", errors.Wrap(err, "malformed ID in the token")
	}
	user, err := m.store.UserInfo(userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user")
	}
	if user == nil {
		return "", errors.Errorf("user not found with ID: %d", userID)
	}

	accessTokens, err := m.store.GetUserAccessTokens(userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user access tokens")
	}

	if !validateAccessToken(accessToken, accessTokens) {
		return "", errors.New("invalid access token")
	}

	log.DebugWithFields("User authenticated",
		log.Fields{"username": user.Uid, "access_token": accessToken})

	return *user.Uid, nil
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

func validateAccessToken(accessTokenString string, userAccessTokens []*store.AccessTokensUserSetting_AccessToken) bool {
	for _, userAccessToken := range userAccessTokens {
		if accessTokenString == userAccessToken.AccessToken {
			return true
		}
	}
	return false
}
