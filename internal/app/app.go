package app

import "github.com/gin-gonic/gin"

type App struct {
	Server  *gin.Engine
	Cleanup func()
}

func (a *App) Run() error {
	// 默认8080端口
	return a.Server.Run(":8080")
}
