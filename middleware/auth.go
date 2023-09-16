package middleware

import (
	"context"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

var (
	ctx              = context.Background()
	db               = model.Db
	rdb              = model.Rdb
	middlewareLogger = log.Log
	excludePath      = []string{"/login", "/register"}
)

// Auth deal with the token passed in
// if token is nil,or can't pass the check,or timeout,it will return error
// query expire time in redis
// TODO: todo 3个auth函数
func Auth(c *gin.Context) {
	// if the path is in excludePath,then skip the check
	//curPath := c.Request.URL.Path
	//if checkExcludePath(curPath) {
	//	return
	//}

	token := c.GetHeader("Authorization")
	// judge if token is nil
	if token == "" {
		c.JSON(http.StatusUnauthorized, result.Failed(result.AuthIncomingTokenFail))
		return
	}
	// parseToken
	claims, parseErr := util.ParseToken(token)
	if parseErr != nil {
		c.JSON(http.StatusUnauthorized, result.Failed(result.AuthParseTokenFail))
		return
	}
	//get username(StudentEmail) and set token info to gin.Context
	username, tokenErr := util.GetUsername(token)
	if tokenErr != nil || username == "" {
		c.JSON(http.StatusUnauthorized, result.Failed(result.AuthError))
		return
	}
	c.Set("username", username)
	c.Set("tokenType", claims.TokenType)
	c.Set("token", token)
	//get Token type,decide the following process
	switch claims.TokenType {

	}

	//refresh token
	//_, err = rdb.Get(ctx, "TOKEN:"+username).Result()
	//if err != nil {
	//	rdb.Set(ctx, "TOKEN:"+username, token, time.Hour*6)
	//}
	// query user in database
	var user model.User
	dbErr := model.Db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	if dbErr != nil {
		// if the user is not exist
		if dbErr == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.UserNotExist))
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthParseTokenFail))
	}
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
