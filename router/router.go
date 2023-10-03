package router

import (
	"net/http"

	v1 "github.com/NJUPT-SAST/sast-link-backend/api/v1"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	var midLog = log.Log
	r.Use(middleware.MiddlewareLogging(midLog))
	// FIXME: need discuss on web log
	// r.Use(middleware.WebLogger)
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
		println("router:" + c.FullPath())
	})

	apiV1 := r.Group("/api/v1")

	usergroup := apiV1.Group("/user")
	{
		usergroup.GET("/info", v1.UserInfo)
		usergroup.POST("/register", v1.Register)
		usergroup.POST("/login", v1.Login)
		usergroup.POST("/logout", v1.Logout)
		usergroup.POST("/changePassword", v1.ChangePassword)
		usergroup.POST("/resetPassword", v1.ResetPassword)
	}
	verify := apiV1.Group("/verify")
	{
		verify.GET("/account", v1.VerifyAccount)
		verify.POST("/captcha", v1.CheckVerifyCode)
	}
	apiV1.GET("/sendEmail", v1.SendEmail)
	//S-LYPL7 admingroup := apiV1.Group("/admin")
	// {
	// }

	// oauth
	oauth := apiV1.Group("/oauth2")
	{
		// authorize
		oauth.GET("/authorize", v1.Authorize)
		// login
		// oauth.GET("/auth", v1.UserAuth)
		oauth.POST("/token", v1.AccessToken)
		oauth.POST("/refresh", v1.RefreshToken)
		oauth.POST("/create-client", v1.CreateClient)
		oauth.GET("/userinfo", v1.OauthUserInfo)
	}

	// third party login
	login := apiV1.Group("/login")
	{
		login.GET("/lark", v1.OauthLarkLogin)
		login.GET("/lark/callback", v1.OauthLarkCallback)
		// login.POST("/github", v1.OauthGithubLogin)
		// login.POST("/qq", v1.OauthQQLogin)
	}
	return r
}
