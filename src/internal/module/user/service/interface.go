package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

// UserService defines the interface for user business logic
type UserService interface {
	// User management
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (*model.User, error)
	ListUsers(ctx context.Context) ([]*model.UserDTO, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// Role management
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error
	CreateRole(ctx context.Context, name, description string) (*model.Role, error)
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, roleID int) error
	GetAllRoles(ctx context.Context) ([]model.Role, error)
	GetUserOrFail(ctx context.Context, userID uuid.UUID) (*model.User, error)
	ConfirmEmail(ctx context.Context, userID uuid.UUID) error

	// Permission management
	AssignPermission(ctx context.Context, userID uuid.UUID, permissionID int) error
	RemovePermission(ctx context.Context, userID uuid.UUID, permissionID int) error
	CreatePermission(ctx context.Context, name, description string) (*model.Permission, error)
	GetAllPermissions(ctx context.Context) ([]model.Permission, error)

	// Authorization
	HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error)

	// User permissions
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]model.Permission, error)
}
