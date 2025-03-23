package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/provider"
	"github.com/xdevspo/go_tmpl_module_app/pkg/jwt"
)

// Константы для ключей контекста
const (
	UserContextKey contextKey = "user"
	RequestKey     contextKey = "request"
)

// Middleware предоставляет middleware для аутентификации
type AuthMiddleware struct {
	jwtManager *jwt.Manager
	sp         provider.ServiceProvider
}

// NewMiddleware создает новый экземпляр middleware для аутентификации
func NewAuthMiddleware(jwtManager *jwt.Manager, sp provider.ServiceProvider) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		sp:         sp,
	}
}

// AuthRequired проверяет наличие и валидность JWT токена
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.sp.Logger().Info("Starting authentication middleware")

		tokenString, err := m.extractToken(c)
		if err != nil {
			m.sp.Logger().WithError(err).Warn("Failed to extract token from request")
			apperrors.ResponseWithError(c, apperrors.UnauthorizedError("errors.unauthorized", err, nil))
			c.Abort()
			return
		}
		m.sp.Logger().WithField("token", tokenString[:10]+"...").Info("Token successfully extracted")

		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			m.sp.Logger().WithError(err).Warn("Token validation failed")
			apperrors.ResponseWithError(c, apperrors.UnauthorizedError("errors.unauthorized", err, nil))
			c.Abort()
			return
		}
		m.sp.Logger().WithField("user_id", claims.UserID).Info("Token validated successfully")

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			m.sp.Logger().WithError(err).Error("Invalid user ID format in token")
			apperrors.ResponseWithError(c, apperrors.UnauthorizedError("errors.invalid_input", err, nil))
			c.Abort()
			return
		}

		user, err := m.sp.UserService(c.Request.Context()).GetByID(c.Request.Context(), userID)
		if err != nil {
			m.sp.Logger().WithError(err).WithField("user_id", userID).Error("Failed to get user by ID")
			apperrors.ResponseWithError(c, err)
			c.Abort()
			return
		}

		if user == nil {
			m.sp.Logger().WithField("user_id", userID).Error("User not found")
			apperrors.ResponseWithError(c, apperrors.NotFoundError("user.not_found", nil, map[string]interface{}{
				"id": userID.String(),
			}))
			c.Abort()
			return
		}

		m.sp.Logger().WithFields(logrus.Fields{
			"user":    user,
			"user_id": user.ID,
			"email":   user.Email,
		}).Info("User successfully retrieved")

		c.Set("user", user)
		c.Set("userId", user.ID)
		c.Set("claims", claims)
		ctx := context.WithValue(c.Request.Context(), UserContextKey, user)
		c.Request = c.Request.WithContext(ctx)

		m.sp.Logger().Info("User data set in context. Auth middleware complete.")

		c.Next()
	}
}

// extractToken извлекает JWT токен из заголовка Authorization
func (m *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("отсутствует заголовок Authorization")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("неверный формат токена")
	}

	return parts[1], nil
}

// RefreshToken проверяет и обновляет токен
func (m *AuthMiddleware) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refreshToken" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.WithValue(c.Request.Context(), RequestKey, c.Request)

		tokenPair, err := m.sp.AuthService(ctx).RefreshToken(ctx, req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен обновления"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"accessToken":     tokenPair.Token,
			"refreshToken":    tokenPair.RefreshToken,
			"accessExpiresAt": tokenPair.AccessExpiresAt,
		})
	}
}

// AuthDebug хендлер для проверки работы авторизации
func (m *AuthMiddleware) AuthDebug() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			apperrors.ResponseWithError(c, apperrors.InternalServerError("errors.internal", nil, map[string]interface{}{
				"message":      "Пользователь не найден в контексте",
				"context_keys": c.Keys,
			}))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Токен действителен",
			"user":         user,
			"context_keys": c.Keys,
		})
	}
}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(ctx context.Context) interface{} {
	if user := ctx.Value(UserContextKey); user != nil {
		return user
	}
	return nil
}
