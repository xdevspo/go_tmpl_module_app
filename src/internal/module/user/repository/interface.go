package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// User methods
	Create(ctx context.Context, user *model.UserModel) error
	Update(ctx context.Context, user *model.UserModel) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.UserModel, error)
	FindByEmail(ctx context.Context, email string) (*model.UserModel, error)
	FindAll(ctx context.Context) ([]*model.UserModel, error)
	ConfirmEmail(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error

	// Role methods
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error
	CreateRole(ctx context.Context, role *model.RoleModel) error
	UpdateRole(ctx context.Context, role *model.RoleModel) error
	DeleteRole(ctx context.Context, roleID int) error
	GetAllRoles(ctx context.Context) ([]model.Role, error)
	FindRoleByName(ctx context.Context, name string) (*model.RoleModel, error)

	// Permission methods
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]model.Permission, error)
	GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error)
	AssignPermission(ctx context.Context, userID uuid.UUID, permissionID int) error
	RemovePermission(ctx context.Context, userID uuid.UUID, permissionID int) error
	CreatePermission(ctx context.Context, permission *model.PermissionModel) error
	UpdatePermission(ctx context.Context, permission *model.PermissionModel) error
	DeletePermission(ctx context.Context, permissionID int) error
	GetAllPermissions(ctx context.Context) ([]model.Permission, error)
	FindPermissionByName(ctx context.Context, name string) (*model.PermissionModel, error)
}
