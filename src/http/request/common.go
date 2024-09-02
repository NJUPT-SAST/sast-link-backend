package request

import (
	"fmt"
	"time"
)

const (
	REGISTER_TICKET_EXP = time.Minute * 5
	RESETPWD_TICKET_EXP = time.Minute * 6
	VERIFY_CODE_EXP     = time.Minute * 3
	OAUTH_TICKET_EXP    = time.Minute * 3
	// This is not login token expire time, this is login ticket expire time
	LOGIN_TICKET_EXP = time.Minute * 5
	// This is login token expire time
	LOGIN_ACCESS_TOKEN_EXP    = time.Hour * 24 * 7
	OAUTH_USER_INFO_EXP = time.Minute * 5

	LARK_CLIENT_TYPE = "lark"
	GITHUB_CLIENT_TYPE = "github"


	// JWT/Redis-key/cookie
	LOGIN_TOKEN_SUB     = "login-token"
	LOGIN_TICKET_SUB    = "login-ticket"
	REGIST_TICKET_SUB   = "register-ticket"
	RESETPWD_TICKET_SUB = "reset-password-ticket"
	OAUTH_LARK_SUB      = "oauth-lark-token"
	OAUTH_GITHUB_SUB    = "oauth-github-token"

	AccessTokenCookieName = "sast-link-access-token"
)

var (
	VERIFY_STATUS = map[string]string{
		"VERIFY_ACCOUNT": "0",
		"SEND_EMAIL":     "1",
		"VERIFY_CAPTCHA": "2",
		"SUCCESS":        "3",
	}
	// Reverse map
	VERIFY_STATUS_REVERSE = map[string]string{
		"0": "VERIFY_ACCOUNT",
		"1": "SEND_EMAIL",
		"2": "VERIFY_CAPTCHA",
		"3": "SUCCESS",
	}
)

// Redis key, for indexing
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

// JWT key
func RegisterJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, REGIST_TICKET_SUB)
}

func ResetPwdJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, RESETPWD_TICKET_SUB)
}

func LoginTicketJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, LOGIN_TICKET_SUB)
}

func LoginJWTSubKey(username string) string {
	return fmt.Sprintf("%s-%s", username, AccessTokenCookieName)
}

func VerifyCodeKey(username string) string {
	return "VerifyCode:" + username
}

// identity is the unique identifier for oauth app user
// like "union_id" for lark, "github_id" for github
func OauthSubKey(identity, oauthType string) string {
	return fmt.Sprintf("%s-%s", identity, oauthType)
}
