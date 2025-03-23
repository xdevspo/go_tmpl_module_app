package handler

import (
	"context"
	"net/http"

	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"

	"github.com/gin-gonic/gin"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/provider"
	"github.com/xdevspo/go_tmpl_module_app/internal/middleware"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/auth/service"
	userModel "github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

type AuthHandler struct {
	authService service.AuthService
	sp          provider.ServiceProvider
}

func NewAuthHandler(authService service.AuthService, sp provider.ServiceProvider) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		sp:          sp,
	}
}

func (h *AuthHandler) handleRegister(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// HandleRegister обрабатывает запрос на регистрацию
func (h *AuthHandler) HandleRegister(c *gin.Context) {
	h.handleRegister(c)
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleLogin обрабатывает запрос на вход в систему
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	h.handleLogin(c)
}

// HandleLogout обрабатывает запрос на выход из системы
func (h *AuthHandler) handleLogout(c *gin.Context) {
	// В будущем здесь можно реализовать инвалидацию токенов
	// или добавление их в черный список

	c.JSON(http.StatusOK, gin.H{
		"message": "успешный выход из системы",
	})
}

// HandleLogout обрабатывает запрос на выход из системы
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	h.handleLogout(c)
}

// Login обрабатывает запрос на вход в систему
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Добавляем HTTP запрос в контекст для получения IP и User-Agent
	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, middleware.RequestKey, c.Request)

	authResponse, err := h.sp.AuthService(ctx).Login(ctx, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// Register обрабатывает запрос на регистрацию
func (h *AuthHandler) Register(c *gin.Context) {
	var req userModel.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Добавляем HTTP запрос в контекст для получения IP и User-Agent
	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, middleware.RequestKey, c.Request)

	authResponse, err := h.sp.AuthService(ctx).Register(ctx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// GetMe возвращает информацию о текущем пользователе
func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден в контексте"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// Logout обрабатывает запрос на выход из системы с отзывом токена
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Добавляем HTTP запрос в контекст для получения IP
	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, middleware.RequestKey, c.Request)

	// Отзываем токен
	err := h.sp.AuthService(ctx).RevokeToken(ctx, req.RefreshToken, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось выйти из системы"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Токен успешно отозван",
	})
}
