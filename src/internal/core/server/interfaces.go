package server

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Server определяет интерфейс для HTTP-сервера
type Server interface {
	Run() error
	Shutdown(ctx context.Context) error
}

type Router interface {
	SetupRouter(engine *gin.Engine)
}

// Handler определяет базовый интерфейс для обработчиков
type Handler interface {
	Register(group *gin.RouterGroup)
}

// Middleware определяет интерфейс для middleware
type Middleware interface {
	Handle() gin.HandlerFunc
}
