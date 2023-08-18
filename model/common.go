package model

import "time"

var (
	REGISTER_TICKET_EXP = time.Minute * 10
	REGISTER_STATUS            = map[string]string{
		"VERIFY_ACCOUNT": "0",
		"SEND_EMAIL":     "1",
		"VERIFY_CAPTCHA": "2",
	}
	CAPTCHA_EXP = time.Minute * 3
	// This is not login token expire time, this is login ticket expire time
	LOGIN_TICKET_EXP = time.Minute * 5
	// This is login token expire time
	LOGIN_TOKEN_EXP = time.Hour * 24 * 7
	// Login token key in redis
	LOGIN_TOKEN_IN_REDIS = "LOGIN"
	
)
