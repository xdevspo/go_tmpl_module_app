package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/config"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/provider"
	"github.com/xdevspo/go_tmpl_module_app/internal/middleware"
	authModel "github.com/xdevspo/go_tmpl_module_app/internal/module/auth/model"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/auth/service"
	userModel "github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
	pkgJwt "github.com/xdevspo/go_tmpl_module_app/pkg/jwt"
)

type authService struct {
	cfg        config.JWTConfig
	sp         provider.ServiceProvider
	jwtManager *pkgJwt.Manager
}

func NewAuthService(cfg config.JWTConfig, sp provider.ServiceProvider, jwtManager *pkgJwt.Manager) service.AuthService {
	return &authService{
		cfg:        cfg,
		sp:         sp,
		jwtManager: jwtManager,
	}
}

func (s *authService) Register(ctx context.Context, req *userModel.CreateUserRequest) (*authModel.AuthResponse, error) {
	createdUser, err := s.sp.UserService(ctx).Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return s.generateToken(ctx, createdUser, getClientIP(ctx))
}

func (s *authService) Login(ctx context.Context, email, password string) (*authModel.AuthResponse, error) {
	userService := s.sp.UserService(ctx)
	user, err := userService.ValidateCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// При успешном входе можно отозвать все существующие refresh токены пользователя
	// или оставить их активными - зависит от требований безопасности
	// s.RevokeAllUserTokens(ctx, user.ID, getClientIP(ctx))

	return s.generateToken(ctx, user, getClientIP(ctx))
}

func (s *authService) generateToken(ctx context.Context, user *userModel.User, ipAddress string) (*authModel.AuthResponse, error) {
	// Вычисляем время истечения токена
	expiresAt := time.Now().Add(time.Duration(s.cfg.AccessTokenExpiryMinutes()) * time.Minute)

	// Роли и разрешения уже должны быть загружены в объекте user
	// Преобразуем роли в строковый массив
	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	// Преобразуем разрешения в строковый массив
	// Включаем как прямые разрешения пользователя, так и разрешения из ролей
	permissionMap := make(map[string]struct{})

	// Добавляем прямые разрешения пользователя
	for _, permission := range user.Permissions {
		permissionMap[permission.Name] = struct{}{}
	}

	// Добавляем разрешения из ролей пользователя
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissionMap[permission.Name] = struct{}{}
		}
	}

	// Преобразуем map в массив
	permissionNames := make([]string, 0, len(permissionMap))
	for name := range permissionMap {
		permissionNames = append(permissionNames, name)
	}

	// Генерируем access token
	accessToken, err := s.jwtManager.GenerateToken(user.ID.String(), roleNames, permissionNames)
	if err != nil {
		return nil, err
	}

	// Генерируем refresh token (с более длительным сроком действия)
	refreshExpiresAt := time.Now().Add(time.Duration(s.cfg.RefreshTokenExpiryHours()) * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"type":    "refresh",
		"exp":     refreshExpiresAt.Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.SecretKey()))
	if err != nil {
		s.sp.Logger().WithError(err).Error("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Сохраняем refresh token в базу данных
	if err := s.saveRefreshToken(ctx, user.ID, refreshTokenString, refreshExpiresAt, ipAddress); err != nil {
		s.sp.Logger().WithError(err).Error("Failed to save refresh token to database")
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &authModel.AuthResponse{
		Token:           accessToken,
		RefreshToken:    refreshTokenString,
		TokenType:       "Bearer",
		ExpiresIn:       int64(s.cfg.AccessTokenExpiryMinutes() * 60), // в секундах
		AccessExpiresAt: expiresAt,
	}, nil
}

// saveRefreshToken сохраняет refresh token в базу данных
func (s *authService) saveRefreshToken(ctx context.Context, userID uuid.UUID, tokenString string, expiresAt time.Time, ipAddress string) error {
	const maxTokensPerUser = 3
	deviceID := getDeviceIdentifier(ctx)

	// Получаем репозиторий
	tokenRepository := s.sp.RefreshTokenRepository(ctx)

	// Проверяем, есть ли уже токен для этого устройства
	existingToken, err := tokenRepository.GetByDeviceIdentifier(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	// Если есть токен для этого устройства, отзываем его
	if existingToken != nil {
		existingToken.Revoke(ipAddress, tokenString)
		if err := tokenRepository.Update(ctx, existingToken); err != nil {
			s.sp.Logger().WithError(err).Error("Failed to revoke existing token for device")
			// Продолжаем, несмотря на ошибку
		}
	}

	// Проверяем, не превышен ли лимит токенов для пользователя
	if err := tokenRepository.RevokeOldestIfLimitExceeded(ctx, userID, maxTokensPerUser-1, ipAddress); err != nil {
		s.sp.Logger().WithError(err).Error("Failed to revoke oldest tokens")
		// Продолжаем, несмотря на ошибку
	}

	refreshToken := &authModel.RefreshToken{
		ID:               uuid.New(),
		UserID:           userID,
		Token:            tokenString,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
		CreatedByIP:      ipAddress,
		DeviceIdentifier: deviceID,
	}

	return tokenRepository.Create(ctx, refreshToken)
}

// getDeviceIdentifier извлекает идентификатор устройства из контекста
func getDeviceIdentifier(ctx context.Context) string {
	// В реальном приложении здесь нужно извлекать идентификатор устройства
	// из HTTP-запроса (User-Agent + другие заголовки) или из специального заголовка
	if req, ok := ctx.Value("request").(*http.Request); ok {
		// Простая реализация на основе User-Agent и IP
		userAgent := req.UserAgent()
		ip := req.RemoteAddr

		// В реальном приложении стоит использовать более надежные методы генерации ID устройства
		// Например, можно использовать куки, локальное хранилище или запрашивать fingerprint устройства
		return fmt.Sprintf("%s-%s", ip, userAgent)
	}

	// Если контекст не содержит запрос, генерируем случайный ID
	return uuid.New().String()
}

// RefreshToken обновляет access token используя refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshTokenString string) (*authModel.AuthResponse, error) {
	const op = "AuthService.RefreshToken"

	// Проверяем токен в базе данных
	tokenRepository := s.sp.RefreshTokenRepository(ctx)
	storedToken, err := tokenRepository.GetByToken(ctx, refreshTokenString)
	if err != nil {
		s.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to find refresh token", op))
		return nil, err
	}

	if storedToken == nil {
		return nil, apperrors.UnauthorizedError("token.not_found", errors.New("refresh token not found"), nil)
	}

	// Проверяем, что токен активен
	if !storedToken.IsActive() {
		return nil, apperrors.UnauthorizedError("token.inactive", errors.New("refresh token is inactive"), nil)
	}

	// Получаем пользователя
	user, err := s.sp.UserService(ctx).GetByID(ctx, storedToken.UserID)
	if err != nil {
		s.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to get user by ID", op))
		return nil, err
	}

	if user == nil {
		return nil, apperrors.NotFoundError("user.not_found", nil, map[string]interface{}{
			"id": storedToken.UserID.String(),
		})
	}

	// Отзываем текущий токен
	ipAddress := getClientIP(ctx)
	// Генерируем новые токены
	authResponse, err := s.generateToken(ctx, user, ipAddress)
	if err != nil {
		return nil, err
	}

	// Помечаем старый токен как использованный
	storedToken.Revoke(ipAddress, authResponse.RefreshToken)
	if err := tokenRepository.Update(ctx, storedToken); err != nil {
		s.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to update refresh token", op))
		// Не возвращаем ошибку, так как новые токены уже созданы
	}

	return authResponse, nil
}

// RevokeToken отзывает указанный refresh токен
func (s *authService) RevokeToken(ctx context.Context, tokenStr string, ipAddress string) error {
	const op = "AuthService.RevokeToken"

	tokenRepository := s.sp.RefreshTokenRepository(ctx)
	token, err := tokenRepository.GetByToken(ctx, tokenStr)
	if err != nil {
		s.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to find refresh token", op))
		return err
	}

	if token == nil {
		return apperrors.NotFoundError("token.not_found", nil, nil)
	}

	token.Revoke(ipAddress, "")
	return tokenRepository.Update(ctx, token)
}

// RevokeAllUserTokens отзывает все токены пользователя
func (s *authService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, ipAddress string) error {
	tokenRepository := s.sp.RefreshTokenRepository(ctx)
	return tokenRepository.RevokeAllUserTokens(ctx, userID, ipAddress)
}

// GetRefreshTokens возвращает все активные токены пользователя
func (s *authService) GetRefreshTokens(ctx context.Context, userID uuid.UUID) ([]authModel.RefreshToken, error) {
	tokenRepository := s.sp.RefreshTokenRepository(ctx)
	return tokenRepository.GetActiveByUserID(ctx, userID)
}

// getClientIP извлекает IP-адрес клиента из контекста
func getClientIP(ctx context.Context) string {
	// В реальном приложении здесь нужно извлекать IP из HTTP-запроса или другого источника
	// Для текущей имплементации используем заглушку
	if req, ok := ctx.Value(middleware.RequestKey).(*http.Request); ok {
		return req.RemoteAddr
	}
	return "unknown"
}
