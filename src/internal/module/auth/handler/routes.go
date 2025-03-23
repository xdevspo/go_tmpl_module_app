package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/xdevspo/go_tmpl_module_app/internal/middleware"
	"github.com/xdevspo/go_tmpl_module_app/pkg/jwt"
)

// RegisterPublicRoutes регистрирует публичные маршруты аутентификации
func (h *AuthHandler) RegisterPublicRoutes(group *gin.RouterGroup) {
	// Публичные маршруты
	group.POST("/register", h.Register)
	group.POST("/login", h.Login)
	group.POST("/refresh", h.RefreshTokenEndpoint)
}

// RegisterProtectedRoutes регистрирует защищенные маршруты аутентификации
func (h *AuthHandler) RegisterProtectedRoutes(group *gin.RouterGroup) {
	// Защищенные маршруты
	group.POST("/logout", h.Logout)
	group.GET("/me", h.GetMe)
	group.GET("/debug", h.AuthDebugEndpoint)
}

// RefreshTokenEndpoint обрабатывает запрос на обновление токена
func (h *AuthHandler) RefreshTokenEndpoint(c *gin.Context) {
	// Получаем JWT конфигурацию из сервис-провайдера
	jwtConfig := h.sp.JWTConfig()
	// Создаем JWT менеджер
	jwtManager := jwt.NewManager(jwtConfig.SecretKey(), jwtConfig.AccessTokenExpiryMinutes())
	// Создаем middleware используя JWT менеджер
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, h.sp)
	authMiddleware.RefreshToken()(c)
}

// AuthDebugEndpoint обрабатывает запрос для проверки авторизации
func (h *AuthHandler) AuthDebugEndpoint(c *gin.Context) {
	// Получаем JWT конфигурацию из сервис-провайдера
	jwtConfig := h.sp.JWTConfig()
	// Создаем JWT менеджер
	jwtManager := jwt.NewManager(jwtConfig.SecretKey(), jwtConfig.AccessTokenExpiryMinutes())
	// Создаем middleware используя JWT менеджер
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, h.sp)
	authMiddleware.AuthDebug()(c)
}
