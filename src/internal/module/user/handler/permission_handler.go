package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/api"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
)

// validatePermissionIDs проверяет существование указанных разрешений и возвращает список несуществующих разрешений
func (h *UserHandler) validatePermissionIDs(c *gin.Context, permissionIDs []int) ([]int, error) {
	if len(permissionIDs) == 0 {
		return nil, apperrors.BadRequestError("errors.invalid_input", nil, map[string]interface{}{
			"message": "список разрешений не может быть пустым",
		})
	}

	// Получаем все доступные разрешения для проверки
	allPermissions, err := h.userService.GetAllPermissions(c.Request.Context())
	if err != nil {
		return nil, err
	}

	// Создаем карту существующих разрешений для быстрого поиска
	existingPermissions := make(map[int]bool)
	for _, perm := range allPermissions {
		existingPermissions[perm.ID] = true
	}

	// Проверяем, что все запрашиваемые разрешения существуют
	var invalidPermissions []int
	for _, permID := range permissionIDs {
		if _, exists := existingPermissions[permID]; !exists {
			invalidPermissions = append(invalidPermissions, permID)
		}
	}

	return invalidPermissions, nil
}

// assignPermissionsHandler назначает разрешения пользователю
func (h *UserHandler) assignPermissionsHandler(c *gin.Context) {
	var req struct {
		PermissionIDs []int `json:"permission_ids" binding:"required"`
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Проверяем наличие всех разрешений
	invalidPermissions, err := h.validatePermissionIDs(c, req.PermissionIDs)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Если найдены несуществующие разрешения, возвращаем ошибку
	if len(invalidPermissions) > 0 {
		apperrors.ResponseWithError(c, apperrors.NotFoundError("permission.not_found", nil, map[string]interface{}{
			"invalid_ids": invalidPermissions,
		}))
		return
	}

	// Добавляем все разрешения
	for _, permID := range req.PermissionIDs {
		if err := h.userService.AssignPermission(c.Request.Context(), userID, permID); err != nil {
			apperrors.ResponseWithError(c, err)
			return
		}
	}

	api.ActionSuccessResponse(c, "response.user.permissions_assigned", gin.H{
		"count": len(req.PermissionIDs),
	})
}

// AssignPermissionsHandler назначает разрешения пользователю
func (h *UserHandler) AssignPermissionsHandler(c *gin.Context) {
	h.assignPermissionsHandler(c)
}

// RevokePermissionHandler отзывает разрешения у пользователя
func (h *UserHandler) RevokePermissionHandler(c *gin.Context) {
	var req struct {
		PermissionIDs []int `json:"permission_ids" binding:"required"`
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Проверяем наличие всех разрешений
	invalidPermissions, err := h.validatePermissionIDs(c, req.PermissionIDs)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Если найдены несуществующие разрешения, возвращаем ошибку
	if len(invalidPermissions) > 0 {
		apperrors.ResponseWithError(c, apperrors.NotFoundError("permission.not_found", nil, map[string]interface{}{
			"invalid_ids": invalidPermissions,
		}))
		return
	}

	// Отзываем все разрешения
	for _, permID := range req.PermissionIDs {
		if err := h.userService.RemovePermission(c.Request.Context(), userID, permID); err != nil {
			apperrors.ResponseWithError(c, err)
			return
		}
	}

	api.ActionSuccessResponse(c, "response.user.permissions_revoked", gin.H{
		"count": len(req.PermissionIDs),
	})
}

// GetUserPermissionsHandler возвращает разрешения пользователя
func (h *UserHandler) GetUserPermissionsHandler(c *gin.Context) {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	permissions, err := h.userService.GetUserPermissions(c.Request.Context(), userId)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	api.SuccessResponse(c, permissions)
}
