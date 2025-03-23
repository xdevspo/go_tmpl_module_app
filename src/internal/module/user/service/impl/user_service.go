package service

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/repository"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/service"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      repository.UserRepository
	logger    logger.Logger
	txManager db.TxManager
}

func NewUserService(repo repository.UserRepository, logger logger.Logger, txManager db.TxManager) service.UserService {
	return &userService{
		repo:      repo,
		logger:    logger,
		txManager: txManager,
	}
}

// Create creates a new user
func (s *userService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	// Validate password confirmation
	if req.Password != req.PasswordConfirmation {
		return nil, apperrors.ValidationError("user.password_mismatch", nil, map[string]any{
			"password":              "***",
			"password_confirmation": "***",
		})
	}

	// Check if user exists
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperrors.ConflictError("user.email_exists", nil, map[string]any{"email": req.Email})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &model.UserModel{
		ID:            uuid.New(),
		Email:         req.Email,
		Password:      string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		MiddleName:    req.MiddleName,
		Phone:         req.Phone,
		Position:      req.Position,
		Active:        req.Active == 1,
		DataRole:      req.DataRole,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	var roles []model.Role
	var permissions []model.Permission

	s.logger.WithFields(logrus.Fields{
		"email":   user.Email,
		"user_id": user.ID,
	}).Info("Starting user creation transaction")

	// Выполняем все операции в одной транзакции
	txErr := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		// Create user
		if err := s.repo.Create(ctx, user); err != nil {
			s.logger.WithError(err).Error("Failed to create user")
			return err
		}

		s.logger.WithField("user_id", user.ID).Info("User created successfully, assigning roles")

		// Перед назначением ролей, проверим, что они существуют
		availableRoles, err := s.repo.GetAllRoles(ctx)
		if err != nil {
			s.logger.WithError(err).Error("Failed to get available roles")
			return err
		}

		s.logger.WithField("available_roles_count", len(availableRoles)).Info("Available roles")
		roleMap := make(map[string]int)
		for _, r := range availableRoles {
			roleMap[r.Name] = r.ID
			s.logger.WithFields(logrus.Fields{
				"role_id":   r.ID,
				"role_name": r.Name,
			}).Info("Available role")
		}

		// Assign roles
		for _, role := range req.Roles {
			s.logger.WithFields(logrus.Fields{
				"user_id":   user.ID,
				"role_name": role.Name,
			}).Info("Looking for role")

			// Сначала проверим в нашем кэше
			if roleID, exists := roleMap[role.Name]; exists {
				s.logger.WithFields(logrus.Fields{
					"user_id":   user.ID,
					"role_id":   roleID,
					"role_name": role.Name,
				}).Info("Assigning permission to user from cache")

				if err := s.repo.AssignRole(ctx, user.ID, roleID); err != nil {
					s.logger.WithError(err).WithField("role", role.Name).Error("Failed to assign role")
					return err
				}
				continue
			}

			// Если не нашли в кэше, ищем в базе данных
			roleModel, err := s.repo.FindRoleByName(ctx, role.Name)
			if err != nil {
				s.logger.WithError(err).WithField("role", role.Name).Error("Failed to find role")
				return err
			}
			if roleModel == nil {
				notFoundErr := apperrors.NotFoundError("role.not_found", nil, map[string]interface{}{"name": role.Name})
				s.logger.WithError(notFoundErr).
					WithField("role", role.Name).
					WithField("errorType", reflect.TypeOf(notFoundErr).String()).
					Error("Role not found")
				return notFoundErr
			}

			s.logger.WithFields(logrus.Fields{
				"user_id":   user.ID,
				"role_id":   roleModel.ID,
				"role_name": roleModel.RoleName,
			}).Info("Assigning role to user")

			if err := s.repo.AssignRole(ctx, user.ID, roleModel.ID); err != nil {
				s.logger.WithError(err).WithField("role", role.Name).Error("Failed to assign role")
				return err
			}
		}

		s.logger.WithField("user_id", user.ID).Info("Roles assigned successfully, assigning permissions")

		// Перед назначением разрешений, проверим, что они существуют
		availablePermissions, err := s.repo.GetAllPermissions(ctx)
		if err != nil {
			s.logger.WithError(err).Error("Failed to get available permissions")
			return err
		}

		s.logger.WithField("available_permissions_count", len(availablePermissions)).Info("Available permissions")
		permMap := make(map[string]int)
		for _, p := range availablePermissions {
			permMap[p.Name] = p.ID
			s.logger.WithFields(logrus.Fields{
				"permission_id":   p.ID,
				"permission_name": p.Name,
			}).Info("Available permission")
		}

		// Assign permissions
		for _, perm := range req.Permissions {
			s.logger.WithFields(logrus.Fields{
				"user_id":         user.ID,
				"permission_name": perm.Name,
			}).Info("Looking for permission")

			// Сначала проверим в нашем кэше
			if permID, exists := permMap[perm.Name]; exists {
				s.logger.WithFields(logrus.Fields{
					"user_id":         user.ID,
					"permission_id":   permID,
					"permission_name": perm.Name,
				}).Info("Assigning permission to user from cache")

				if err := s.repo.AssignPermission(ctx, user.ID, permID); err != nil {
					s.logger.WithError(err).WithField("permission", perm.Name).Error("Failed to assign permission")
					return err
				}
				continue
			}

			// Если не нашли в кэше, ищем в базе данных
			permModel, err := s.repo.FindPermissionByName(ctx, perm.Name)
			if err != nil {
				s.logger.WithError(err).WithField("permission", perm.Name).Error("Failed to find permission")
				return err
			}
			if permModel == nil {
				notFoundErr := apperrors.NotFoundError("permission.not_found", nil, map[string]interface{}{"name": perm.Name})
				s.logger.WithError(notFoundErr).
					WithField("permission", perm.Name).
					WithField("errorType", reflect.TypeOf(notFoundErr).String()).
					WithField("errorInterface", fmt.Sprintf("%#v", notFoundErr)).
					Error("Permission not found")

				// Проверяем, что ошибка относится к типу AppError для гарантии отката транзакции
				s.logger.WithField("errorIsAppError", fmt.Sprintf("%T", notFoundErr)).Info("Checking error type")

				return notFoundErr
			}

			s.logger.WithFields(logrus.Fields{
				"user_id":         user.ID,
				"permission_id":   permModel.ID,
				"permission_name": permModel.PermissionName,
			}).Info("Assigning permission to user")

			if err := s.repo.AssignPermission(ctx, user.ID, permModel.ID); err != nil {
				s.logger.WithError(err).WithField("permission", perm.Name).Error("Failed to assign permission")
				return err
			}
		}

		s.logger.WithField("user_id", user.ID).Info("Permissions assigned successfully, retrieving roles and permissions for response")

		// Получаем роли и разрешения для ответа (все еще в транзакции)
		r, err := s.repo.GetUserRoles(ctx, user.ID)
		if err != nil {
			s.logger.WithError(err).WithField("userId", user.ID).Error("Failed to get user roles")
			return err
		}
		roles = r

		p, err := s.repo.GetUserPermissions(ctx, user.ID)
		if err != nil {
			s.logger.WithError(err).WithField("userId", user.ID).Error("Failed to get user permissions")
			return err
		}
		permissions = p

		s.logger.WithField("user_id", user.ID).Info("Transaction completed successfully")
		return nil
	})

	if txErr != nil {
		s.logger.WithError(txErr).
			WithField("errorType", reflect.TypeOf(txErr).String()).
			Error("Transaction failed")

		// Создаем новый контекст для проверки, т.к. старый контекст может содержать отмененную транзакцию
		newCtx := context.Background()
		// Проверим, остался ли пользователь в базе данных несмотря на ошибку
		checkUser, checkErr := s.repo.FindByEmail(newCtx, user.Email)
		if checkErr != nil {
			s.logger.WithError(checkErr).Error("Failed to check if user was rolled back")
		}

		if checkUser != nil {
			s.logger.WithField("email", user.Email).Error("WARNING: User was not rolled back after transaction error!")
		} else {
			s.logger.WithField("email", user.Email).Info("User was successfully rolled back after transaction error")
		}

		return nil, txErr
	}

	s.logger.WithField("user_id", user.ID).Info("User created with roles and permissions")
	return model.FromDBModel(user, roles, permissions), nil
}

// Update updates an existing user
func (s *userService) Update(ctx context.Context, user *model.User) error {
	dbUser := user.ToDBModel()
	dbUser.UpdatedAt = time.Now()
	return s.repo.Update(ctx, dbUser)
}

// Delete deletes an existing user
func (s *userService) delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetUserOrFail(ctx, id)
	if err != nil {
		return err
	}

	return s.delete(ctx, id)
}

// GetByID gets a user by ID
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	roles, err := s.repo.GetUserRoles(ctx, id)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetUserPermissions(ctx, id)
	if err != nil {
		return nil, err
	}

	return model.FromDBModel(user, roles, permissions), nil
}

// GetByEmail gets a user by email
func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return model.FromDBModel(user, roles, permissions), nil
}

// ValidateCredentials validates user credentials
func (s *userService) ValidateCredentials(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.NotFoundError("user.not_found", err, map[string]interface{}{"email": email})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperrors.UnauthorizedError("errors.invalid_credentials", err, nil)
	}

	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return model.FromDBModel(user, roles, permissions), nil
}

// ListUsers возвращает список всех пользователей без ролей и разрешений
func (s *userService) ListUsers(ctx context.Context) ([]*model.UserDTO, error) {
	s.logger.WithField("component", "UserService.ListUsersLight").Debug("Getting all users")

	// Получаем список пользователей из репозитория
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get users from repository")
		return nil, err
	}

	// Создаем слайс для моделей пользователей без ролей и разрешений
	result := make([]*model.UserDTO, 0, len(users))

	// Преобразуем каждую модель из БД в DTO
	for _, dbUser := range users {
		// Преобразуем в DTO и добавляем в результат
		user := model.UserDTOFromDBModel(dbUser)
		result = append(result, user)
	}

	return result, nil
}

// AssignRole assigns a role to a user
func (s *userService) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	return s.repo.AssignRole(ctx, userID, roleID)
}

// RemoveRole removes a role from a user
func (s *userService) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	return s.repo.RemoveRole(ctx, userID, roleID)
}

// CreateRole creates a new role
func (s *userService) CreateRole(ctx context.Context, name, description string) (*model.Role, error) {
	role := &model.RoleModel{
		RoleName:    name,
		Description: description,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return &model.Role{
		ID:          role.ID,
		Name:        role.RoleName,
		Description: role.Description,
	}, nil
}

// UpdateRole updates an existing role
func (s *userService) UpdateRole(ctx context.Context, role *model.Role) error {
	dbRole := &model.RoleModel{
		ID:          role.ID,
		RoleName:    role.Name,
		Description: role.Description,
	}
	return s.repo.UpdateRole(ctx, dbRole)
}

// DeleteRole deletes an existing role
func (s *userService) DeleteRole(ctx context.Context, roleID int) error {
	return s.repo.DeleteRole(ctx, roleID)
}

// GetAllRoles gets all roles
func (s *userService) GetAllRoles(ctx context.Context) ([]model.Role, error) {
	return s.repo.GetAllRoles(ctx)
}

// AssignPermission Permission management
func (s *userService) assignPermission(ctx context.Context, userID uuid.UUID, permissionID int) error {
	return s.repo.AssignPermission(ctx, userID, permissionID)
}

func (s *userService) AssignPermission(ctx context.Context, userID uuid.UUID, permissionID int) error {
	_, err := s.GetUserOrFail(ctx, userID)
	if err != nil {
		return err
	}
	return s.assignPermission(ctx, userID, permissionID)
}

// RemovePermission Permission management
func (s *userService) RemovePermission(ctx context.Context, userID uuid.UUID, permissionID int) error {
	return s.repo.RemovePermission(ctx, userID, permissionID)
}

// CreatePermission Permission management
func (s *userService) CreatePermission(ctx context.Context, name, description string) (*model.Permission, error) {
	permission := &model.PermissionModel{
		PermissionName: name,
		Description:    description,
	}

	if err := s.repo.CreatePermission(ctx, permission); err != nil {
		return nil, err
	}

	return &model.Permission{
		ID:          permission.ID,
		Name:        permission.PermissionName,
		Description: permission.Description,
	}, nil
}

// GetAllPermissions Permission management
func (s *userService) GetAllPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.repo.GetAllPermissions(ctx)
}

// HasRole checks if a user has a given role
func (s *userService) HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}
	return false, nil
}

// HasPermission checks if a user has a given permission
func (s *userService) HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	permissions, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Name == permissionName {
			return true, nil
		}
	}
	return false, nil
}

// GetUserRoles gets all roles for a user
func (s *userService) getUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *userService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	_, err := s.GetUserOrFail(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.getUserRoles(ctx, userID)
}

// GetUserPermissions получает разрешения пользователя
func (s *userService) getUserPermissions(ctx context.Context, userID uuid.UUID) ([]model.Permission, error) {
	return s.repo.GetUserPermissions(ctx, userID)
}

func (s *userService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]model.Permission, error) {
	_, err := s.GetUserOrFail(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.getUserPermissions(ctx, userID)
}

// ConfirmEmail confirms a user's email
func (s *userService) confirmEmail(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ConfirmEmail(ctx, userID)
}

func (s *userService) ConfirmEmail(ctx context.Context, userID uuid.UUID) error {
	_, err := s.GetUserOrFail(ctx, userID)
	if err != nil {
		return err
	}

	return s.confirmEmail(ctx, userID)
}

// GetUserOrFail возвращает пользователя по ID или ошибку, если пользователя нет
func (s *userService) GetUserOrFail(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	// Запрашиваем пользователя из БД
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperrors.NotFoundError("user.not_found", nil, map[string]interface{}{
			"id": userID,
		})
	}
	return user, nil
}

// ChangePassword изменяет пароль пользователя
func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	s.logger.WithField("component", "UserService.ChangePassword").
		WithField("user_id", userID).
		Debug("Changing user password")

	// Получаем пользователя из БД
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return apperrors.NotFoundError("user.not_found", nil, map[string]interface{}{
			"id": userID,
		})
	}

	// Проверяем старый пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return apperrors.BadRequestError("user.invalid_password", err, map[string]interface{}{
			"message": "Неверный текущий пароль",
		})
	}

	// Проверяем, что новый пароль соответствует требованиям
	if len(newPassword) < 8 {
		return apperrors.ValidationError("user.password_too_short", nil, map[string]interface{}{
			"min_length": 8,
		})
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.InternalServerError("errors.internal", err, nil)
	}

	// Сохраняем новый пароль
	return s.repo.ChangePassword(ctx, userID, string(hashedPassword))
}
