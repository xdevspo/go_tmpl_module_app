package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/api"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
)

// validateRoleIDs проверяет существование указанных ролей и возвращает список несуществующих ролей
func (h *UserHandler) validateRoleIDs(c *gin.Context, roleIDs []int) ([]int, error) {
	if len(roleIDs) == 0 {
		return nil, apperrors.BadRequestError("errors.invalid_input", nil, map[string]interface{}{
			"message": "список ролей не может быть пустым",
		})
	}

	// Получаем все доступные роли для проверки
	allRoles, err := h.userService.GetAllRoles(c.Request.Context())
	if err != nil {
		return nil, err
	}

	// Создаем карту существующих ролей для быстрого поиска
	existingRoles := make(map[int]bool)
	for _, role := range allRoles {
		existingRoles[role.ID] = true
	}

	// Проверяем, что все запрашиваемые роли существуют
	var invalidRoles []int
	for _, roleID := range roleIDs {
		if _, exists := existingRoles[roleID]; !exists {
			invalidRoles = append(invalidRoles, roleID)
		}
	}

	return invalidRoles, nil
}

// assignRoleHandler назначает роли пользователю
func (h *UserHandler) assignRoleHandler(c *gin.Context) {
	var req struct {
		RoleIDs []int `json:"role_ids" binding:"required"`
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

	// Проверяем наличие всех ролей
	invalidRoles, err := h.validateRoleIDs(c, req.RoleIDs)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Если найдены несуществующие роли, возвращаем ошибку
	if len(invalidRoles) > 0 {
		apperrors.ResponseWithError(c, apperrors.NotFoundError("role.not_found", nil, map[string]interface{}{
			"invalid_ids": invalidRoles,
		}))
		return
	}

	// Добавляем все роли
	for _, roleID := range req.RoleIDs {
		if err := h.userService.AssignRole(c.Request.Context(), userID, roleID); err != nil {
			apperrors.ResponseWithError(c, err)
			return
		}
	}

	api.ActionSuccessResponse(c, "response.user.roles_assigned", gin.H{
		"count": len(req.RoleIDs),
	})
}

// AssignRoleHandler назначает роль пользователю
func (h *UserHandler) AssignRoleHandler(c *gin.Context) {
	h.assignRoleHandler(c)
}

// revokeRoleHandler отзывает роли у пользователя
func (h *UserHandler) revokeRoleHandler(c *gin.Context) {
	var req struct {
		RoleIDs []int `json:"role_ids" binding:"required"`
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

	// Проверяем наличие всех ролей
	invalidRoles, err := h.validateRoleIDs(c, req.RoleIDs)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	// Если найдены несуществующие роли, возвращаем ошибку
	if len(invalidRoles) > 0 {
		apperrors.ResponseWithError(c, apperrors.NotFoundError("role.not_found", nil, map[string]interface{}{
			"invalid_ids": invalidRoles,
		}))
		return
	}

	// Отзываем все роли
	for _, roleID := range req.RoleIDs {
		if err := h.userService.RemoveRole(c.Request.Context(), userID, roleID); err != nil {
			apperrors.ResponseWithError(c, err)
			return
		}
	}

	api.ActionSuccessResponse(c, "response.user.roles_revoked", gin.H{
		"count": len(req.RoleIDs),
	})
}

// RevokeRoleHandler отзывает роль у пользователя
func (h *UserHandler) RevokeRoleHandler(c *gin.Context) {
	h.revokeRoleHandler(c)
}

// GetUserRolesHandler возвращает роли пользователя
func (h *UserHandler) GetUserRolesHandler(c *gin.Context) {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		apperrors.ResponseWithError(c, apperrors.BadRequestError("errors.invalid_input", err, map[string]interface{}{
			"message": "неверный формат ID",
		}))
		return
	}

	roles, err := h.userService.GetUserRoles(c.Request.Context(), userId)
	if err != nil {
		apperrors.ResponseWithError(c, err)
		return
	}

	api.SuccessResponse(c, roles)
}
