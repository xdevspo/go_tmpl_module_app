package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
	modelUser "github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/repository"
)

// userRepository реализует интерфейс repository.UserRepository
type userRepository struct {
	db        db.Client
	txManager db.TxManager
	logger    logger.Logger
}

func NewRepository(db db.Client, txManager db.TxManager, logger logger.Logger) repository.UserRepository {
	return &userRepository{
		db:        db,
		txManager: txManager,
		logger:    logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *modelUser.UserModel) error {
	q := db.Query{
		Name: "user.Create",
		QueryRaw: `
			INSERT INTO users (
				id, email, password, first_name, last_name, middle_name,
				phone, position, active, data_role, email_verified,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11,
				$12, $13
			)
		`,
	}
	_, err := r.db.DB().ExecContext(ctx, q,
		user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.MiddleName,
		user.Phone, user.Position, user.Active, user.DataRole, user.EmailVerified,
		user.CreatedAt, user.UpdatedAt,
	)

	return err
}

func (r *userRepository) Update(ctx context.Context, user *modelUser.UserModel) error {
	q := db.Query{
		Name:     "user.Update",
		QueryRaw: `UPDATE users SET email = $2, first_name = $3, last_name = $4 WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, user.ID, user.Email, user.FirstName, user.LastName)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := db.Query{
		Name:     "user.Delete",
		QueryRaw: `UPDATE users SET deleted_at = NOW() WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, id)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*modelUser.UserModel, error) {
	q := db.Query{
		Name:     "user.FindByID",
		QueryRaw: `SELECT id, email, password, first_name, last_name FROM users WHERE id = $1 AND deleted_at IS NULL`,
	}
	var user modelUser.UserModel
	err := r.db.DB().ScanOneContext(ctx, &user, q, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*modelUser.UserModel, error) {
	q := db.Query{
		Name:     "user.FindByEmail",
		QueryRaw: `SELECT id, email, password, first_name, last_name FROM users WHERE email = $1 AND deleted_at IS NULL`,
	}
	var user modelUser.UserModel
	err := r.db.DB().ScanOneContext(ctx, &user, q, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]*modelUser.UserModel, error) {

	q := db.Query{
		Name: "user.FindAll",
		QueryRaw: `
			SELECT id, email, password, first_name, last_name, middle_name,
				phone, position, active, data_role, email_verified,
				created_at, updated_at, last_login, deleted_at
			FROM users 
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
		`,
	}

	var users []*modelUser.UserModel
	err := r.db.DB().ScanAllContext(ctx, &users, q)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Role methods
func (r *userRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]modelUser.Role, error) {
	// Получаем роли пользователя
	roleQuery := db.Query{
		Name:     "user.GetUserRoles",
		QueryRaw: `SELECT r.id, r.role_name as name, r.description FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1`,
	}
	var roles []modelUser.Role
	err := r.db.DB().ScanAllContext(ctx, &roles, roleQuery, userID)
	if err != nil {
		return nil, err
	}

	// Для каждой роли загружаем разрешения
	for i := range roles {
		permQuery := db.Query{
			Name: "user.GetRolePermissions",
			QueryRaw: `
				SELECT p.id, p.permission_name as name, p.description
				FROM permissions p
				JOIN role_permissions rp ON rp.permission_id = p.id
				WHERE rp.role_id = $1
			`,
		}
		var permissions []modelUser.Permission
		err := r.db.DB().ScanAllContext(ctx, &permissions, permQuery, roles[i].ID)
		if err != nil {
			return nil, err
		}
		roles[i].Permissions = permissions
	}

	return roles, nil
}

func (r *userRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	q := db.Query{
		Name:     "user.AssignRole",
		QueryRaw: `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, userID, roleID)
	return err
}

func (r *userRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	q := db.Query{
		Name:     "user.RemoveRole",
		QueryRaw: `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, userID, roleID)
	return err
}

func (r *userRepository) CreateRole(ctx context.Context, role *modelUser.RoleModel) error {
	q := db.Query{
		Name:     "user.CreateRole",
		QueryRaw: `INSERT INTO roles (role_name, description) VALUES ($1, $2)`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, role.RoleName, role.Description)
	return err
}

func (r *userRepository) UpdateRole(ctx context.Context, role *modelUser.RoleModel) error {
	q := db.Query{
		Name:     "user.UpdateRole",
		QueryRaw: `UPDATE roles SET role_name = $2, description = $3 WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, role.ID, role.RoleName, role.Description)
	return err
}

func (r *userRepository) DeleteRole(ctx context.Context, roleID int) error {
	q := db.Query{
		Name:     "user.DeleteRole",
		QueryRaw: `DELETE FROM roles WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, roleID)
	return err
}

func (r *userRepository) GetAllRoles(ctx context.Context) ([]modelUser.Role, error) {
	q := db.Query{
		Name:     "user.GetAllRoles",
		QueryRaw: `SELECT id, role_name as name, description FROM roles`,
	}
	var roles []modelUser.Role
	err := r.db.DB().ScanAllContext(ctx, &roles, q)
	return roles, err
}

// Permission methods
func (r *userRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]modelUser.Permission, error) {
	// Используем UNION для объединения разрешений из двух источников
	q := db.Query{
		Name: "user.GetUserPermissions",
		QueryRaw: `
			-- Прямые разрешения пользователя
			SELECT DISTINCT p.id, p.permission_name as name, p.description
			FROM permissions p
			JOIN user_permissions up ON p.id = up.permission_id
			WHERE up.user_id = $1
			
			UNION
			
			-- Разрешения, полученные через роли
			SELECT DISTINCT p.id, p.permission_name as name, p.description
			FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = $1
		`,
	}

	var permissions []modelUser.Permission
	err := r.db.DB().ScanAllContext(ctx, &permissions, q, userID)

	return permissions, err
}

func (r *userRepository) GetRolePermissions(ctx context.Context, roleID int) ([]modelUser.Permission, error) {
	q := db.Query{
		Name: "user.GetRolePermissions",
		QueryRaw: `
			SELECT p.id, p.permission_name as name, p.description
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			WHERE rp.role_id = $1
		`,
	}
	var permissions []modelUser.Permission
	err := r.db.DB().ScanAllContext(ctx, &permissions, q, roleID)
	return permissions, err
}

func (r *userRepository) AssignPermission(ctx context.Context, userID uuid.UUID, permissionID int) error {
	q := db.Query{
		Name:     "user.AssignPermission",
		QueryRaw: `INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, userID, permissionID)
	return err
}

func (r *userRepository) RemovePermission(ctx context.Context, userID uuid.UUID, permissionID int) error {
	q := db.Query{
		Name:     "user.RemovePermission",
		QueryRaw: `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, userID, permissionID)
	return err
}

func (r *userRepository) CreatePermission(ctx context.Context, permission *modelUser.PermissionModel) error {
	q := db.Query{
		Name:     "user.CreatePermission",
		QueryRaw: `INSERT INTO permissions (permission_name, description) VALUES ($1, $2)`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, permission.PermissionName, permission.Description)
	return err
}

func (r *userRepository) UpdatePermission(ctx context.Context, permission *modelUser.PermissionModel) error {
	q := db.Query{
		Name:     "user.UpdatePermission",
		QueryRaw: `UPDATE permissions SET permission_name = $2, description = $3 WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, permission.ID, permission.PermissionName, permission.Description)
	return err
}

func (r *userRepository) DeletePermission(ctx context.Context, permissionID int) error {
	q := db.Query{
		Name:     "user.DeletePermission",
		QueryRaw: `DELETE FROM permissions WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, permissionID)
	return err
}

func (r *userRepository) GetAllPermissions(ctx context.Context) ([]modelUser.Permission, error) {
	q := db.Query{
		Name:     "user.GetAllPermissions",
		QueryRaw: `SELECT id, permission_name as name, description FROM permissions`,
	}
	var permissions []modelUser.Permission
	err := r.db.DB().ScanAllContext(ctx, &permissions, q)
	return permissions, err
}

func (r *userRepository) FindRoleByName(ctx context.Context, name string) (*modelUser.RoleModel, error) {
	q := db.Query{
		Name:     "user.FindRoleByName",
		QueryRaw: `SELECT id, role_name, description FROM roles WHERE role_name = $1`,
	}
	var role modelUser.RoleModel
	err := r.db.DB().ScanOneContext(ctx, &role, q, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *userRepository) FindPermissionByName(ctx context.Context, name string) (*modelUser.PermissionModel, error) {
	q := db.Query{
		Name:     "user.FindPermissionByName",
		QueryRaw: `SELECT id, permission_name, description FROM permissions WHERE permission_name = $1`,
	}
	var permission modelUser.PermissionModel
	err := r.db.DB().ScanOneContext(ctx, &permission, q, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *userRepository) ConfirmEmail(ctx context.Context, userID uuid.UUID) error {
	q := db.Query{
		Name:     "user.ConfirmEmail",
		QueryRaw: `UPDATE users SET email_verified = true WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, userID)

	return err
}

func (r *userRepository) ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	q := db.Query{
		Name:     "user.ChangePassword",
		QueryRaw: `UPDATE users SET password = $2 WHERE id = $1`,
	}
	_, err := r.db.DB().ExecContext(ctx, q, userID, newPassword)

	return err
}
