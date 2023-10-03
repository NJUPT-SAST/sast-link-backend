package middleware

import (
	"context"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	ctx              = context.Background()
	db               = model.Db
	rdb              = model.Rdb
	middlewareLogger = log.Log
	excludePath      = []string{"/login", "/register"}
)

// RegisterAuth Auth with register process,Token type:Register-TICKET
func RegisterAuth(c *gin.Context) {
	//verify token parse
	username, tokenType, token, checkRes := checkToken(c)
	if !checkRes.Success {
		c.AbortWithStatusJSON(http.StatusUnauthorized, checkRes)
		return
	}
	//verify tokenType
	if tokenType != model.REGIST_TICKET_SUB {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthTokenTypeError))
		return
	}
	//get Register ticket value
	status, err := model.Rdb.Get(c, token).Result()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.CheckTicketNotfound))
		log.Log.WithFields(logrus.
			Fields{
			"username": username,
		}).Error(err)
	}
	//judge if the Register process is correct by router and val
	switch c.Request.URL.Path {
	case "/api/v1/sendEmail":
		// Determine if the ticket is correct
		if status != model.REGISTER_STATUS["VERIFY_ACCOUNT"] {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.TicketNotCorrect))
			return
		}
	case "/api/v1/verify/captcha":
		// judge if the ticket is correct
		if status != model.REGISTER_STATUS["SEND_EMAIL"] {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.TicketNotCorrect))
			return
		}
	case "/api/v1/user/register":
		// check which phase current in
		switch status {
		case model.REGISTER_STATUS["VERIFY_ACCOUNT"], model.REGISTER_STATUS["SEND_EMAIL"]:
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.RegisterPhaseError))
			return
		case model.REGISTER_STATUS["SUCCESS"]:
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.UserAlreadyExist))
			return
		}
	}
	//refresh token
	//_, err = rdb.Get(ctx, "TOKEN:"+username).Result()
	//if err != nil {
	//	rdb.Set(ctx, "TOKEN:"+username, token, time.Hour*6)
	//}

	// query user in database
	//var user model.User
	//dbErr := model.Db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	//if dbErr != nil {
	//	// if the user is not exist
	//	if dbErr == gorm.ErrRecordNotFound {
	//		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.UserNotExist))
	//	}
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthParseTokenFail))
	//}

}

// LoginAuth The router of user/admin operations use this Auth,Token type:Login-TOKEN
func LoginAuth(c *gin.Context) {
	_, tokenType, _, checkRes := checkToken(c)
	if !checkRes.Success {
		c.AbortWithStatusJSON(http.StatusUnauthorized, checkRes)
		return
	}
	//verify tokenType
	if tokenType != model.LOGIN_TICKET_SUB {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthTokenTypeError))
		return
	}

}

// checkToken Beginning of the whole Auth Process
func checkToken(c *gin.Context) (string, string, string, result.Response) {
	token := c.GetHeader("Authorization")
	// judge if token is nil
	if token == "" {
		return "", "", "", result.Failed(result.AuthIncomingTokenFail)
	}
	// parseToken
	claims, parseErr := util.ParseToken(token)
	if parseErr != nil {
		return "", "", "", result.Failed(result.AuthParseTokenFail)
	}
	//get username(StudentEmail) and set token info to gin.Context
	username, tokenErr := util.GetUsername(token)
	if tokenErr != nil || username == "" {
		return "", "", "", result.Failed(result.AuthError)
	}
	c.Set("username", username)
	c.Set("tokenType", claims.TokenType)
	c.Set("token", token)
	return username, claims.TokenType, token, result.Success(nil)
}

func checkExcludePath(path string) bool {
	for _, p := range excludePath {
		if p == path {
			return true
		}
	}
	return false
}

// checkTokenInRedis check TICKET or TOKEN in redis
// return if checkTokenInRedis PASS
// `tokenPattern` const of "TICKET" or "TOKEN"
func checkTokenInRedis(username string, tokenPattern string) (bool, error) {
	_, err := rdb.Get(ctx, tokenPattern+":"+username).Result()
	// todo replace here by logging
	if err != nil {
		return false, err
	}
	return true, nil
}
