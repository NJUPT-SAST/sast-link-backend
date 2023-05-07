package router

import (
	"net/http"

	v1 "github.com/NJUPT-SAST/sast-link-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
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
	}
	verify := apiV1.Group("/verify")
	{
		verify.GET("/account", v1.VerifyAccount)
	}
	apiV1.GET("/sendEmail", v1.SendEmail)
	//S-LYPL7 admingroup := apiV1.Group("/admin")
	// {
	// }

	return r
}
