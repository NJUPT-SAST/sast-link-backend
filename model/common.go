package model

import (
	"fmt"
	"time"
)

const (
	REGISTER_TICKET_EXP = time.Minute * 5
	RESETPWD_TICKET_EXP = time.Minute * 6
	CAPTCHA_EXP         = time.Minute * 3
	// This is not login token expire time, this is login ticket expire time
	LOGIN_TICKET_EXP = time.Minute * 5
	// This is login token expire time
	LOGIN_TOKEN_EXP = time.Hour * 24 * 7
	// Login token key in redis
	LOGIN_TOKEN_IN_REDIS = "LOGIN"

	// For JWT
	LOGIN_TOKEN_SUB     = "loginToken"
	LOGIN_TICKET_SUB    = "loginTicket"
	REGIST_TICKET_SUB   = "registerTicket"
	RESETPWD_TICKET_SUB = "resetPwdTicket"
)

var (
	VERIFY_STATUS = map[string]string{
		"VERIFY_ACCOUNT": "0",
		"SEND_EMAIL":     "1",
		"VERIFY_CAPTCHA": "2",
		"SUCCESS":        "3",
	}
)

func RegisterTicketKey(ticket string) string {
	return "REGISTER_TICKET:" + ticket
}

func LoginTicketKey(username string) string {
	return "LOGIN_TICKET:" + username
}

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
	return fmt.Sprintf("%s-%s", username, LOGIN_TOKEN_SUB)
}

func LoginTokenKey(username string) string {
	return "TOKEN:" + username
}

func CaptchaKey(username string) string {
	return "CAPTCHA:" + username
}
