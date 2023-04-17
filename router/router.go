package router

import (
	"net/http"

	v1 "github.com/NJUPT-SAST/sast-link-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	apiV1 := r.Group("/api/v1")

	usergroup := apiV1.Group("/user")
	{
		usergroup.POST("/register", v1.Register)
	}

	// admingroup := apiV1.Group("/admin")
	// {
	// }

	return r
}
