package request

import "net/http"

type ContextKey int

const (
	ClientIPContextKey ContextKey = iota
	UserIDContextKey
	UserNameContextKey
	UserRolesContextKey
	IsAuthenticatedContextKey
    AccessTokenContextKey
)

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}

// GetUserID returns the user ID from the request context.
//
// The user ID is set by the AuthenticationInterceptor middleware.
// The user ID is the database primary key of the user and not the student ID.
func GetUserID(r *http.Request) string {
	return getContextStringValue(r, UserIDContextKey)
}

// GetUsername returns the username from the request context.
//
// The username is set by the AuthenticationInterceptor middleware.
// The username is the student ID of the user.
func GetUsername(r *http.Request) string {
	return getContextStringValue(r, UserNameContextKey)
}

// GetIsAuthenticated returns the authentication status from the request context.
//
// The authentication status is set by the AuthenticationInterceptor middleware.
func GetIsAuthenticated(r *http.Request) bool {
	if v := r.Context().Value(IsAuthenticatedContextKey); v != nil {
		if value, valid := v.(bool); valid {
			return value
		}
	}
	return false
}

func GetAccessToken(r *http.Request) string {
	return getContextStringValue(r, AccessTokenContextKey)
}

// func GetUserRole(r *http.Request) {
// }
