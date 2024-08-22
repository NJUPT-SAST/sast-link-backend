package middleware

import "strings"

var authenticationAllowlist = map[string]bool{
	"/api/v1/login":               true,
	"/api/v1/register":            true,
	"/api/v1/check_verify_code":   true,
	"/api/v1/user/reset_password": true,
	"/api/v1/verify/*":            true,
	"/api/v1/oauth2/*":            true,
	"/api/v1/sendEmail":           true,
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

var allowedPathOnlyForAdmin = map[string]bool{}

// isOnlyForAdminAllowedPath returns true if the method is allowed to be called only by admin.
func isOnlyForAdminAllowedPath(methodName string) bool {
	return allowedPathOnlyForAdmin[methodName]
}
