package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	container "github.com/xdevspo/go_tmpl_module_app/internal/core/container"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/i18n"
	corepolicy "github.com/xdevspo/go_tmpl_module_app/internal/core/policy"
	"github.com/xdevspo/go_tmpl_module_app/internal/middleware"
	authHandlers "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/handler"
	userHandlers "github.com/xdevspo/go_tmpl_module_app/internal/module/user/handler"
	userPolicy "github.com/xdevspo/go_tmpl_module_app/internal/module/user/policy"
	"github.com/xdevspo/go_tmpl_module_app/pkg/jwt"
)

// SetupRouter configures all routes and middleware for the HTTP server
func SetupRouter(ctx context.Context, sp *container.ServiceProvider) *gin.Engine {
	logger := sp.Logger()
	authService := sp.AuthService(ctx)

	translator := i18n.GetInstance()
	if err := translator.LoadTranslations("internal/core/i18n/translations"); err != nil {
		logger.WithError(err).Error("Failed to load translations")
	}

	router := gin.Default()

	router.Use(middleware.RequestLoggerWithLogger(logger))

	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"services": "auth-services",
		})
	})

	authHandler := authHandlers.NewAuthHandler(authService, sp)

	userService := sp.UserService(ctx)
	userHandler := userHandlers.NewUserHandler(userService, sp)

	jwtConfig := sp.JWTConfig()
	jwtManager := jwt.NewManager(jwtConfig.SecretKey(), jwtConfig.AccessTokenExpiryMinutes())
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, sp)

	policyFactory := corepolicy.NewPolicyFactory()

	registerModulePolicies(policyFactory)

	policyMiddleware := middleware.NewPolicyMiddleware(policyFactory)

	apiV1 := router.Group("/api/v1")

	auth := apiV1.Group("/auth")
	authHandler.RegisterPublicRoutes(auth)

	authProtected := auth.Group("")
	authProtected.Use(authMiddleware.Authenticate())
	authHandler.RegisterProtectedRoutes(authProtected)

	usersGroup := apiV1.Group("/users")
	usersGroup.Use(authMiddleware.Authenticate())

	// Регистрируем маршруты
	userHandler.RegisterUserRoutes(usersGroup, policyMiddleware)
	userHandler.RegisterUserPermissionRoutes(usersGroup, policyMiddleware)
	userHandler.RegisterUserRoleRoutes(usersGroup, policyMiddleware)

	return router
}

// registerModulePolicies регистрирует политики всех модулей в центральной фабрике
func registerModulePolicies(factory *corepolicy.PolicyFactory) {
	userPolicy.RegisterInFactory(factory)

	// Здесь можно добавить регистрацию политик других модулей
	// somemodule.RegisterInFactory(factory)
}
