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
	"gorm.io/gorm"
)

var (
	ctx              = context.Background()
	db               = model.Db
	rdb              = model.Rdb
	middlewareLogger = log.Log
	excludePath      = []string{"/login", "/register"}
)

// JWT deal with the token passed in
// if token is nil,or can't pass the check,or timeout,it will return error
// query expire time in redis
// TODO: refresh jwt expire time in redis
func JWT(c *gin.Context) {
	curPath := c.Request.URL.Path
	// if the path is in excludePath,then skip the check
	if checkExcludePath(curPath) {
		return
	}

	token := c.GetHeader("TOKEN")
	// token is nil
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthParseTokenFail))
	}
	claims, err := util.ParseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthParseTokenFail))
	}
	username, claimsError := claims.GetSubject()
	if claimsError != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, result.Failed(result.AuthParseTokenFail))
	}
	//refresh token
	_, err = rdb.Get(ctx, "TOKEN:"+username).Result()
	if err != nil {
		rdb.Set(ctx, "TOKEN:"+username, token, time.Hour*6)
	}
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
