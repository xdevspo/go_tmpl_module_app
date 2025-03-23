package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/api"
	container "github.com/xdevspo/go_tmpl_module_app/internal/core/container"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/service"
)

// UserHandler обрабатывает HTTP-запросы для модуля пользователей
type UserHandler struct {
	userService service.UserService
	sp          *container.ServiceProvider
}

// NewUserHandler создаёт новый экземпляр UserHandler
func NewUserHandler(userService service.UserService, sp *container.ServiceProvider) *UserHandler {
	return &UserHandler{
		userService: userService,
		sp:          sp,
	}
}

// ChangePassword изменяет пароль пользователя
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Получаем ID пользователя из URL параметра
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	// Вызываем метод сервиса для изменения пароля
	if err := h.userService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	api.ActionSuccessResponse(c, "response.user.password_changed", nil)
}

// ListUsers возвращает список пользователей (для админа)
func (h *UserHandler) ListUsers(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		apperrors.ResponseWithError(c, apperrors.UnauthorizedError("errors.unauthorized", nil, nil))
		return
	}

	// Проверяем наличие прав на основе ролей или разрешений
	currentUser, ok := user.(*model.User)
	if !ok {
		apperrors.ResponseWithError(c, apperrors.InternalServerError("errors.internal", nil, nil))
		return
	}

	// Проверка на роли admin или full, или разрешение users:read
	hasAccess := currentUser.HasRole("admin") || currentUser.HasRole("full") || currentUser.HasPermission("users:read")
	if !hasAccess {
		apperrors.ResponseWithError(c, apperrors.ForbiddenError("errors.forbidden", nil, map[string]interface{}{
			"message": "недостаточно прав для просмотра списка пользователей",
		}))
		return
	}

	users, err := h.userService.ListUsers(c.Request.Context())
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	api.SuccessResponse(c, users)
}

// GetUserByID возвращает пользователя по ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	// Проверяем, есть ли пользователь в контексте и совпадает ли его ID
	contextUserObj, exists := c.Get("user")
	var user *model.User

	if exists {
		if contextUser, ok := contextUserObj.(*model.User); ok && contextUser.ID == userId {
			// Используем пользователя из контекста, если его ID совпадает с запрашиваемым
			h.sp.Logger().WithField("user_id", userId).Debug("Using user from context instead of DB query")
			user = contextUser
		}
	}

	// Если пользователь не найден в контексте, запрашиваем из БД
	if user == nil {
		h.sp.Logger().WithField("user_id", userId).Debug("User not found in context, querying DB")
		var err error
		user, err = h.userService.GetByID(c.Request.Context(), userId)
		if err != nil {
			apperrors.ResponseWithError(c, err)
			return
		}

		if user == nil {
			apperrors.ResponseWithError(c, apperrors.NotFoundError("user.not_found", nil, map[string]interface{}{
				"id": id,
			}))
			return
		}
	}

	api.SuccessResponse(c, user)
}

// DeleteUser удаляет пользователя
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	if err := h.userService.Delete(c.Request.Context(), userId); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	api.ActionSuccessResponse(c, "response.user.deleted", nil)
}

// VerifyEmail подтверждает email пользователя
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	if err := h.userService.ConfirmEmail(c.Request.Context(), userId); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	api.ActionSuccessResponse(c, "response.user.email_confirmed", gin.H{"user_id": userId})
}
