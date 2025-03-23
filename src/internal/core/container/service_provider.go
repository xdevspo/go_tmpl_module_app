package app

import (
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db/pg"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db/transaction"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/closer"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/config"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
	authRepo "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/repository"
	authRepoImpl "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/repository/impl"
	authService "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/service"
	authServiceImpl "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/service/impl"
	userRepo "github.com/xdevspo/go_tmpl_module_app/internal/module/user/repository"
	userRepoPostgres "github.com/xdevspo/go_tmpl_module_app/internal/module/user/repository/postgres"
	userService "github.com/xdevspo/go_tmpl_module_app/internal/module/user/service"
	userServiceImpl "github.com/xdevspo/go_tmpl_module_app/internal/module/user/service/impl"
	"github.com/xdevspo/go_tmpl_module_app/pkg/jwt"
)

type ServiceProvider struct {
	appConfig  config.AppConfig
	pgConfig   config.PGConfig
	jwtConfig  config.JWTConfig
	httpConfig config.HTTPConfig

	logrusLogger *logrus.Logger
	logger       logger.Logger

	dbClient  db.Client
	txManager db.TxManager

	userRepository         userRepo.UserRepository
	refreshTokenRepository authRepo.RefreshTokenRepository

	userService userService.UserService
	authService authService.AuthService
}

func NewServiceProvider() *ServiceProvider {
	sp := &ServiceProvider{}

	sp.logrusLogger = sp.createDefaultLogger()
	sp.logger = logger.NewLogrusAdapter(sp.logrusLogger)

	return sp
}

// createDefaultLogger creates a logger with default settings
func (sp *ServiceProvider) createDefaultLogger() *logrus.Logger {
	logger := logrus.New()

	shouldLog := false

	env := os.Getenv("APP_ENV")
	if env == "dev" {
		shouldLog = true
	} else {
		logEnabled := os.Getenv("APP_LOG")
		shouldLog = logEnabled == "true"
	}

	if shouldLog {
		if env == "dev" {
			logger.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
			})
		} else {
			logger.SetFormatter(&logrus.JSONFormatter{})
		}

		logger.SetOutput(os.Stdout)

		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "info"
		}

		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			level = logrus.InfoLevel
		}

		logger.SetLevel(level)

		logger.WithFields(logrus.Fields{
			"environment": env,
			"log_level":   level.String(),
		}).Info("Logging initialized")
	} else {
		logger.SetOutput(io.Discard)
		logger.SetLevel(logrus.PanicLevel)
	}

	return logger
}

// Logger возвращает настроенный логгер
func (sp *ServiceProvider) Logger() logger.Logger {
	return sp.logger
}

// LogrusLogger возвращает оригинальный logrus логгер (для обратной совместимости)
func (sp *ServiceProvider) LogrusLogger() *logrus.Logger {
	return sp.logrusLogger
}

// ConfigureLogger configures the logger with the specified parameters
func (sp *ServiceProvider) ConfigureLogger(level logrus.Level, useJSON bool) {
	sp.logrusLogger.SetLevel(level)

	if useJSON {
		sp.logrusLogger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		sp.logrusLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

func (sp *ServiceProvider) AppConfig() config.AppConfig {
	if sp.appConfig == nil {
		cfg, err := config.NewAppConfig()
		if err != nil {
			sp.logger.Fatalf("failed to get app config: %s", err.Error())
		}
		sp.appConfig = cfg
	}
	return sp.appConfig
}

func (sp *ServiceProvider) PGConfig() config.PGConfig {
	if sp.pgConfig == nil {
		cfg, err := config.NewPGConfig()
		if err != nil {
			sp.logger.Fatalf("failed to get pg config: %s", err.Error())
		}

		sp.pgConfig = cfg
	}

	return sp.pgConfig
}

func (sp *ServiceProvider) JWTConfig() config.JWTConfig {
	if sp.jwtConfig == nil {
		cfg, err := config.NewJWTConfig()
		if err != nil {
			sp.logger.Fatalf("failed to get jwt config: %s", err.Error())
		}

		sp.jwtConfig = cfg
	}

	return sp.jwtConfig
}

func (sp *ServiceProvider) HTTPConfig() config.HTTPConfig {
	if sp.httpConfig == nil {
		cfg, err := config.NewHTTPConfig()
		if err != nil {
			sp.logger.Fatalf("failed to get http config: %s", err.Error())
		}

		sp.httpConfig = cfg
	}

	return sp.httpConfig
}

func (sp *ServiceProvider) DBClient(ctx context.Context) db.Client {
	if sp.dbClient == nil {
		dbClient, err := pg.New(ctx, sp.PGConfig().DSN(), sp.Logger())
		if err != nil {
			sp.logger.Fatalf("failed to create db client: %v", err)
		}

		err = dbClient.DB().Ping(ctx)
		if err != nil {
			sp.logger.Fatalf("failed to ping db: %v", err)
		}
		closer.Add(dbClient.Close)

		sp.dbClient = dbClient
	}

	return sp.dbClient
}

func (sp *ServiceProvider) TxManager(ctx context.Context) db.TxManager {
	if sp.txManager == nil {
		sp.txManager = transaction.NewTransactionManager(sp.DBClient(ctx).DB())
	}

	return sp.txManager
}

func (sp *ServiceProvider) UserRepository(ctx context.Context) userRepo.UserRepository {
	if sp.userRepository == nil {
		sp.userRepository = userRepoPostgres.NewRepository(sp.DBClient(ctx), sp.TxManager(ctx), sp.Logger())
	}

	return sp.userRepository
}

func (sp *ServiceProvider) RefreshTokenRepository(ctx context.Context) authRepo.RefreshTokenRepository {
	if sp.refreshTokenRepository == nil {
		sp.refreshTokenRepository = authRepoImpl.NewRefreshTokenRepository(sp, sp.DBClient(ctx).DB())
	}
	return sp.refreshTokenRepository
}

func (sp *ServiceProvider) UserService(ctx context.Context) userService.UserService {
	if sp.userService == nil {
		sp.userService = userServiceImpl.NewUserService(sp.UserRepository(ctx), sp.Logger(), sp.TxManager(ctx))
	}
	return sp.userService
}

func (sp *ServiceProvider) AuthService(ctx context.Context) authService.AuthService {
	if sp.authService == nil {
		jwtConfig := sp.JWTConfig()
		jwtManager := jwt.NewManager(jwtConfig.SecretKey(), jwtConfig.AccessTokenExpiryMinutes())
		sp.authService = authServiceImpl.NewAuthService(jwtConfig, sp, jwtManager)
	}
	return sp.authService
}
