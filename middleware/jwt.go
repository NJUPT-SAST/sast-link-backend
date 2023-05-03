package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
)

var ctx = context.Background()
var rdb = model.Rdb

// JWT deal with the token passed in
// if token is nil,or can`t pass the check,or timeout,it will return error
// query expire time in redis
// todo refresh jwt expire time in redis
func JWT(c *gin.Context) {
	var code = result.SUCCESS
	var data any

	if token := c.Query("token"); token == "" {
		// judge if token is nil
		code = result.INVALID_PARAMS
	} else {
		// judge if the Token is valid
		if claims, err := util.ParseToken(token); err != nil {
			code = result.ERROR_AUTH_CHECK_TOKEN_FAIL
		} else {
			// verify token in redis
			// notice: there are two ways to verify,one is to verify TICKET, another is TOKEN
			username, claimsError := claims.GetSubject()
			if claimsError != nil {
				log.Log.Errorf("Parse Token Error")
				return
			}
			//here to add the router where api need judge ticket in redis
			if path := c.FullPath(); path == "/login" {
				checkTokenInRedis(username, "TICKET", &code)
			} else {
				//here the api need judge token in redis
				//if Pass and Token time<=1hour need to refresh
				if checkTokenInRedis(username, "TOKEN", &code) {
					prefixToken := "TOKEN" + ":" + username
					if expireTime := rdb.ExpireTime(ctx, prefixToken); expireTime.Val().Hours() <= 0.5 {
						err := rdb.Set(ctx, prefixToken, token, 3*time.Hour).Err()
						// todo replace here by logging
						if err != nil {
							log.Log.Errorf("Refresh TOKEN error: %v", err)
						}
					}
				}
			}
		}
	}

	if code != result.SUCCESS {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": code,
			"msg":  result.GetMsg(code),
			"data": data,
		})
		c.Abort()
		return
	}
}

// checkTokenInRedis check TICKET or TOKEN in redis
// return if checkTokenInRedis PASS
//
// `tokenPattern` const of "TICKET" or "TOKEN"
func checkTokenInRedis(username string, tokenPattern string, code *int) bool {
	val, err := rdb.Get(ctx, tokenPattern+":"+username).Result()
	// todo replace here by logging
	if err != nil {
		log.Log.Errorf("Check %s in Redis error: %v", tokenPattern, err)
	}
	if val == "" {
		*code = result.ERROR_AUTH_CHECK_TICKET_NOTFOUND
	}
	return *code == result.SUCCESS
}
