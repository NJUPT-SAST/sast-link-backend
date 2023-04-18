package middleware

import (
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/gin-gonic/gin"
)

func WebLogger(ctx *gin.Context) {
	log.Log.Infof(`
		RequestURI:		%s
		RequestMethod: 	%s
		Handler:		%s
	`,
		ctx.Request.URL,
		ctx.Request.Method,
	)
}
