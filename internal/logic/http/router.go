package http

import (
	"github.com/gin-gonic/gin"
)

type baseRouter struct {
	G *gin.RouterGroup
}

func initRouter() *gin.Engine {
	r := gin.Default()
	g := r.Group("/logic")
	{
		v1(g)
	}
	return r
}

func v1(router *gin.RouterGroup) {
	r := router.Group("/v1")
	r.GET("/push/key")
}
