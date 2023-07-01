package model

import "time"

var (
	REGISTER_TICKET_EXPIRE_TIME = time.Minute * 10
	REGISTER_STATUS            = map[string]string{
		"VERIFY_ACCOUNT": "0",
		"SEND_EMAIL":     "1",
		"VERIFY_CAPTCHA": "2",
	}
	CAPTCHA_EXPIRE_TIME = time.Minute * 3
	// This is not login token expire time, this is login ticket expire time
	LOGIN_TICKET_EXPIRE = time.Minute * 5
)
