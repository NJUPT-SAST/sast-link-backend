package middleware

import "strings"

// Don't need to check the authentication for the following paths.
var authenticationAllowlist = map[string]bool{
	"/api/v1/user/login":         true,
	"/api/v1/user/register":      true,
	"/api/v1/user/resetPassword": true,
	"/api/v1/user/loginWithSSO":  true,
	"/api/v1/verify/*":           true,
	"/api/v1/oauth2/authorize":   true,
	"/api/v1/oauth2/token":       true,
	"/api/v1/oauth2/refresh":     true,
	"/api/v1/oauth2/userinfo":    true,
	"/api/v1/sendEmail":          true,
	"/api/v1/listIDP":            true,
	"/api/v1/idpInfo":            true,
}

// isUnauthorizeAllowed returns whether the method is exempted from authentication.
// Support the wildcard character *.
func isUnauthorizeAllowed(fullMethodName string) bool {
	for k := range authenticationAllowlist {
		if strings.HasSuffix(k, "*") {
			if strings.HasPrefix(fullMethodName, strings.TrimSuffix(k, "*")) {
				return true
			}
		}
	}

	return authenticationAllowlist[fullMethodName]
}

//nolint
var allowedPathOnlyForAdmin = map[string]bool{}

//nolint
// isOnlyForAdminAllowedPath returns true if the method is allowed to be called only by admin.
func isOnlyForAdminAllowedPath(methodName string) bool {
	return allowedPathOnlyForAdmin[methodName]
}
