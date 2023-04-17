package v1

import (
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/gin-gonic/gin"
)

func Register(ctx *gin.Context) {
	// TODO: fill relevant code
	ctx.JSON(http.StatusOK, result.Success())
}
