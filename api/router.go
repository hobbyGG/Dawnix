package api

import "github.com/gin-gonic/gin"

type RouterRegistrar interface {
	// 定义路由接口方法
	Register(r *gin.RouterGroup)
}

func NewRouter(middlewares []gin.HandlerFunc, registrars ...RouterRegistrar) *gin.Engine {
	r := gin.Default()
	if len(middlewares) > 0 {
		r.Use(middlewares...)
	}
	apiV1Group := r.Group("api/v1")

	for _, registrar := range registrars {
		registrar.Register(apiV1Group)
	}
	return r
}
