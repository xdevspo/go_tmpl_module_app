package model

import (
	"time"

	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// UserModel represents the user in the database
type UserModel struct {
	ID            uuid.UUID        `db:"id"`
	Email         string           `db:"email"`
	Password      string           `db:"password"`
	FirstName     string           `db:"first_name"`
	LastName      string           `db:"last_name"`
	MiddleName    string           `db:"middle_name"`
	Phone         string           `db:"phone"`
	Position      string           `db:"position"`
	Active        bool             `db:"active"`
	DataRole      string           `db:"data_role"`
	EmailVerified bool             `db:"email_verified"`
	LastLogin     pgtype.Timestamp `db:"last_login"`
	CreatedAt     time.Time        `db:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at"`
}

// RoleModel represents the role in the database
type RoleModel struct {
	ID          int    `db:"id"`
	RoleName    string `db:"role_name"`
	Description string `db:"description"`
}

// PermissionModel represents the permission in the database
type PermissionModel struct {
	ID             int    `db:"id"`
	PermissionName string `db:"permission_name"`
	Description    string `db:"description"`
}

// IsEmpty checks if the user is empty
func (um *UserModel) IsEmpty() bool {
	if um == nil {
		return true
	}
	return um.ID == uuid.Nil
}

// BeforeCreate generates UUID
func (um *UserModel) BeforeCreate() {
	if um.ID == uuid.Nil {
		um.ID = uuid.New()
	}
}

// BeforeSave prepares the password for hashing
func (um *UserModel) BeforeSave() error {
	if len(um.Password) > 0 && len(um.Password) < 60 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(um.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		um.Password = string(hashedPassword)
	}
	return nil
}

// ComparePassword compares passwords
func (um *UserModel) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(um.Password), []byte(password))
	return err == nil
}

// FullName returns the full name
func (um *UserModel) FullName() string {
	return um.FirstName + " " + um.LastName
}

// User represents the business model with roles and permissions
type User struct {
	ID            uuid.UUID    `json:"id"`
	Email         string       `json:"email"`
	Password      string       `json:"password"`
	FirstName     string       `json:"firstName"`
	LastName      string       `json:"lastName"`
	MiddleName    string       `json:"middleName"`
	Phone         string       `json:"phone"`
	Position      string       `json:"position"`
	Active        bool         `json:"active"`
	DataRole      string       `json:"dataRole"`
	EmailVerified bool         `json:"emailVerified"`
	LastLogin     *time.Time   `json:"lastLogin,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
	DeletedAt     *time.Time   `json:"deletedAt,omitempty"`
	Roles         []Role       `json:"roles,omitempty"`
	Permissions   []Permission `json:"permissions,omitempty"`
}

// Role represents the business model for role
type Role struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// Permission represents the business model for permission
type Permission struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UserDTO представляет модель пользователя для API (без ролей и разрешений)
// DTO (Data Transfer Object) - объект для передачи данных через API
type UserDTO struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	MiddleName    string     `json:"middleName"`
	Phone         string     `json:"phone"`
	Position      string     `json:"position"`
	Active        bool       `json:"active"`
	DataRole      string     `json:"dataRole"`
	EmailVerified bool       `json:"emailVerified"`
	LastLogin     *time.Time `json:"lastLogin,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ToDBModel converts business model to database model
func (u *User) ToDBModel() *UserModel {
	var lastLogin pgtype.Timestamp
	if u.LastLogin != nil {
		lastLogin.Time = *u.LastLogin
		lastLogin.Status = pgtype.Present
	}

	var deletedAt pgtype.Timestamp
	if u.DeletedAt != nil {
		deletedAt.Time = *u.DeletedAt
		deletedAt.Status = pgtype.Present
	}

	return &UserModel{
		ID:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		MiddleName:    u.MiddleName,
		Phone:         u.Phone,
		Position:      u.Position,
		Active:        u.Active,
		DataRole:      u.DataRole,
		EmailVerified: u.EmailVerified,
		LastLogin:     lastLogin,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

// FromDBModel creates business model from database model
func FromDBModel(dbUser *UserModel, roles []Role, permissions []Permission) *User {
	var lastLogin *time.Time
	if dbUser.LastLogin.Status == pgtype.Present {
		lastLogin = &dbUser.LastLogin.Time
	}

	var deletedAt *time.Time
	if dbUser.DeletedAt.Status == pgtype.Present {
		deletedAt = &dbUser.DeletedAt.Time
	}

	return &User{
		ID:            dbUser.ID,
		Email:         dbUser.Email,
		Password:      dbUser.Password,
		FirstName:     dbUser.FirstName,
		LastName:      dbUser.LastName,
		MiddleName:    dbUser.MiddleName,
		Phone:         dbUser.Phone,
		Position:      dbUser.Position,
		Active:        dbUser.Active,
		DataRole:      dbUser.DataRole,
		EmailVerified: dbUser.EmailVerified,
		LastLogin:     lastLogin,
		CreatedAt:     dbUser.CreatedAt,
		UpdatedAt:     dbUser.UpdatedAt,
		DeletedAt:     deletedAt,
		Roles:         roles,
		Permissions:   permissions,
	}
}

// HasRole checks if user has specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission checks if user has specific permission
func (u *User) HasPermission(permissionName string) bool {
	// Check direct permissions
	for _, perm := range u.Permissions {
		if perm.Name == permissionName {
			return true
		}
	}

	// Check permissions through roles
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm.Name == permissionName {
				return true
			}
		}
	}

	return false
}

// HasAnyPermission checks if user has any of the given permissions
func (u *User) HasAnyPermission(permissions ...string) bool {
	return slices.ContainsFunc(permissions, u.HasPermission)
}

// UserDTOFromDBModel создает объект передачи данных (DTO) из модели базы данных
func UserDTOFromDBModel(dbUser *UserModel) *UserDTO {
	var lastLogin *time.Time
	if dbUser.LastLogin.Status == pgtype.Present {
		lastLogin = &dbUser.LastLogin.Time
	}

	return &UserDTO{
		ID:            dbUser.ID,
		Email:         dbUser.Email,
		FirstName:     dbUser.FirstName,
		LastName:      dbUser.LastName,
		MiddleName:    dbUser.MiddleName,
		Phone:         dbUser.Phone,
		Position:      dbUser.Position,
		Active:        dbUser.Active,
		DataRole:      dbUser.DataRole,
		EmailVerified: dbUser.EmailVerified,
		LastLogin:     lastLogin,
		CreatedAt:     dbUser.CreatedAt,
		UpdatedAt:     dbUser.UpdatedAt,
	}
}
