package provider

import (
	"context"

	"github.com/xdevspo/go_tmpl_module_app/internal/core/config"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
	authRepo "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/repository"
	authService "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/service"
	userRepo "github.com/xdevspo/go_tmpl_module_app/internal/module/user/repository"
	userService "github.com/xdevspo/go_tmpl_module_app/internal/module/user/service"
)

// ServiceProvider defines the interface for accessing services
type ServiceProvider interface {
	Logger() logger.Logger
	AppConfig() config.AppConfig
	JWTConfig() config.JWTConfig
	HTTPConfig() config.HTTPConfig
	UserRepository(ctx context.Context) userRepo.UserRepository
	UserService(ctx context.Context) userService.UserService
	AuthService(ctx context.Context) authService.AuthService
	RefreshTokenRepository(ctx context.Context) authRepo.RefreshTokenRepository
}
