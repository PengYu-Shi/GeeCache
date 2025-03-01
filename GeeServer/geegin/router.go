package geegin

import (
	"GeeServer/geegin/router"
	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine) *gin.Engine {
	user := r.Group("/user")
	user.GET("/gee", router.Get)
	user.POST("/gee", router.Get)

	return r
}
