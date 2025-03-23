package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/xdevspo/go_tmpl_module_app/internal/middleware"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/policy"
)

// RegisterRoutes регистрирует все маршруты модуля пользователей
func (h *UserHandler) RegisterRoutes(group *gin.RouterGroup, policyMiddleware *middleware.PolicyMiddleware) {
	h.RegisterUserRoutes(group, policyMiddleware)
	h.RegisterUserRoleRoutes(group, policyMiddleware)
	h.RegisterUserPermissionRoutes(group, policyMiddleware)
}

// RegisterUserRoutes регистрирует маршруты для управления пользователями
func (h *UserHandler) RegisterUserRoutes(group *gin.RouterGroup, policyMiddleware *middleware.PolicyMiddleware) {
	group.POST("/:id/change-password", h.ChangePassword)
	group.GET("/:id/verify-email", h.VerifyEmail)

	group.GET("", policyMiddleware.RequirePermission(policy.ResourceName, "view"), h.ListUsers)
	group.GET("/:id", policyMiddleware.RequirePermission(policy.ResourceName, "view"), h.GetUserByID)
	group.DELETE("/:id", policyMiddleware.RequirePermission(policy.ResourceName, "delete"), h.DeleteUser)
}

// RegisterUserRoleRoutes регистрирует маршруты для управления ролями
func (h *UserHandler) RegisterUserRoleRoutes(group *gin.RouterGroup, policyMiddleware *middleware.PolicyMiddleware) {
	group.POST("/:id/roles", policyMiddleware.RequirePermission(policy.ResourceName, "assign-role"), h.AssignRoleHandler)
	group.DELETE("/:id/roles", policyMiddleware.RequirePermission(policy.ResourceName, "revoke-role"), h.RevokeRoleHandler)
	group.GET("/:id/roles", policyMiddleware.RequirePermission(policy.ResourceName, "view-roles"), h.GetUserRolesHandler)
}

// RegisterUserPermissionRoutes регистрирует маршруты для управления разрешениями
func (h *UserHandler) RegisterUserPermissionRoutes(group *gin.RouterGroup, policyMiddleware *middleware.PolicyMiddleware) {
	group.POST("/:id/permissions", policyMiddleware.RequirePermission(policy.ResourceName, "assign-permission"), h.AssignPermissionsHandler)
	group.DELETE("/:id/permissions", policyMiddleware.RequirePermission(policy.ResourceName, "revoke-permission"), h.RevokePermissionHandler)
	group.GET("/:id/permissions", policyMiddleware.RequirePermission(policy.ResourceName, "view-permissions"), h.GetUserPermissionsHandler)
}
