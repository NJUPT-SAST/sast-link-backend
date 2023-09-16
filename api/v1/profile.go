package v1

import "github.com/gin-gonic/gin"

func GetProfile(ctx *gin.Context) {
	ctx.JSON(200, "success")
}
func CreateProfile(ctx *gin.Context) {
	ctx.JSON(200, "success")
}
func UpdateProfile(ctx *gin.Context) {
	ctx.JSON(200, "success")
}
func UploadAvatar(ctx *gin.Context) {
	ctx.JSON(200, "success")
}
