package request

import (
	"fmt"
	"time"
)

const (
	RegisterTicketExp = time.Minute * 5
	ResetPwdTicketExp = time.Minute * 6
	VerifyCodeExp     = time.Minute * 3
	OAuthTicketExp    = time.Minute * 3
	BindEmailExp      = time.Minute * 5
	// This is not login token expire time, this is login ticket expire time.
	LoginTicketExp = time.Minute * 5
	// This is login token expire time.
	LoginAccessTokenExp = time.Hour * 24 * 7
	OAuthUserInfoExp    = time.Minute * 5

	LarkClientType   = "lark"
	GithubClientType = "github"

	// Used for JWT/Redis-key/cookie.
	LoginTokenSub      = "login-token"
	LoginTicketSub     = "login-ticket"
	RegisterTicketSub  = "register-ticket"
	ResetPwdTicketSub  = "reset-password-ticket"
	OAuthCheckEmailSub = "oauth-check-email-ticket"

	AccessTokenCookieName = "sast-link-access-token"
)

var (
	VerifyStatus = map[string]string{
		"VERIFY_ACCOUNT": "0",
		"SEND_EMAIL":     "1",
		"VERIFY_CAPTCHA": "2",
		"SUCCESS":        "3",
	}
	// Reverse map.
	VerifyStatusReverse = map[string]string{
		"0": "VERIFY_ACCOUNT",
		"1": "SEND_EMAIL",
		"2": "VERIFY_CAPTCHA",
		"3": "SUCCESS",
	}
)

// Redis key, for indexing.
func RegisterTicketKey(ticket string) string {
	return "REGISTER_TICKET:" + ticket
}

func LoginTicketKey(username string) string {
	return "LOGIN_TICKET:" + username
}

func LoginTokenKey(username string) string {
	return "ACCESS_TOKEN:" + username
}

func CaptchaKey(username string) string {
	return "CAPTCHA:" + username
}

// JWT key.
func RegisterJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, RegisterTicketSub)
}

func ResetPwdJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, ResetPwdTicketSub)
}

func LoginTicketJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, LoginTicketSub)
}

func LoginJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, AccessTokenCookieName)
}

func BindSSOSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, OAuthCheckEmailSub)
}

func VerifyCodeKey(username string) string {
	return "VerifyCode:" + username
}

// identity is the unique identifier for oauth app user
// like "union_id" for lark, "github_id" for github.
func OauthSubKey(identity, oauthType string) string {
	return fmt.Sprintf("%s-%s", identity, oauthType)
}
