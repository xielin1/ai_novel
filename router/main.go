package router

import (
	"embed"
	"github.com/gin-gonic/gin"
)

func SetRouter(router *gin.Engine, buildFS embed.FS, indexPage []byte, controllers *APIControllers) {
	SetApiRouter(router, controllers)
	setWebRouter(router, buildFS, indexPage)
}
