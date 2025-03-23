package impl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/client/db"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/provider"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/auth/model"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/auth/repository"
)

type refreshTokenRepository struct {
	sp   provider.ServiceProvider
	db   db.DB
	name string
}

// NewRefreshTokenRepository создает новый экземпляр репозитория для refresh токенов
func NewRefreshTokenRepository(sp provider.ServiceProvider, db db.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{
		sp:   sp,
		db:   db,
		name: "RefreshTokenRepository",
	}
}

// Create создает новый refresh токен в базе данных
func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	const op = "RefreshTokenRepository.Create"
	if token == nil {
		return apperrors.InternalServerError("token.is_nil", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".Create",
		QueryRaw: `
			INSERT INTO refresh_tokens 
			(id, user_id, token, expires_at, created_at, created_by_ip, device_identifier)
			VALUES 
			($1, $2, $3, $4, $5, $6, $7)
		`,
	}

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, q,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
		token.CreatedByIP,
		token.DeviceIdentifier,
	)
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to create refresh token", op))
		return apperrors.InternalServerError("token.create_error", err, nil)
	}

	return nil
}

// GetByToken находит токен по его значению
func (r *refreshTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*model.RefreshToken, error) {
	const op = "RefreshTokenRepository.GetByToken"
	if tokenStr == "" {
		return nil, apperrors.BadRequestError("token.empty", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".GetByToken",
		QueryRaw: `
			SELECT id, user_id, token, expires_at, revoked, created_at, created_by_ip,
			       revoked_at, revoked_by_ip, replaced_by_token, device_identifier
			FROM refresh_tokens
			WHERE token = $1
		`,
	}

	row := r.db.QueryRowContext(ctx, q, tokenStr)
	var token model.RefreshToken

	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.Revoked,
		&token.CreatedAt,
		&token.CreatedByIP,
		&token.RevokedAt,
		&token.RevokedByIP,
		&token.ReplacedByToken,
		&token.DeviceIdentifier,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to get refresh token", op))
		return nil, apperrors.InternalServerError("token.get_error", err, nil)
	}

	return &token, nil
}

// GetActiveByUserID находит все активные токены пользователя
func (r *refreshTokenRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.RefreshToken, error) {
	const op = "RefreshTokenRepository.GetActiveByUserID"
	if userID == uuid.Nil {
		return nil, apperrors.BadRequestError("user_id.empty", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".GetActiveByUserID",
		QueryRaw: `
			SELECT id, user_id, token, expires_at, revoked, created_at, created_by_ip,
			       revoked_at, revoked_by_ip, replaced_by_token, device_identifier
			FROM refresh_tokens
			WHERE user_id = $1 AND revoked = false AND expires_at > $2
		`,
	}

	rows, err := r.db.QueryContext(ctx, q, userID, time.Now())
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to get active refresh tokens", op))
		return nil, apperrors.InternalServerError("token.get_error", err, nil)
	}
	defer rows.Close()

	var tokens []model.RefreshToken
	for rows.Next() {
		var token model.RefreshToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.ExpiresAt,
			&token.Revoked,
			&token.CreatedAt,
			&token.CreatedByIP,
			&token.RevokedAt,
			&token.RevokedByIP,
			&token.ReplacedByToken,
			&token.DeviceIdentifier,
		)
		if err != nil {
			r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to scan refresh token", op))
			return nil, apperrors.InternalServerError("token.scan_error", err, nil)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: rows error", op))
		return nil, apperrors.InternalServerError("token.rows_error", err, nil)
	}

	return tokens, nil
}

// Update обновляет информацию о токене
func (r *refreshTokenRepository) Update(ctx context.Context, token *model.RefreshToken) error {
	const op = "RefreshTokenRepository.Update"
	if token == nil {
		return apperrors.InternalServerError("token.is_nil", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".Update",
		QueryRaw: `
			UPDATE refresh_tokens
			SET revoked = $1, revoked_at = $2, revoked_by_ip = $3, replaced_by_token = $4
			WHERE id = $5
		`,
	}

	_, err := r.db.ExecContext(ctx, q,
		token.Revoked,
		token.RevokedAt,
		token.RevokedByIP,
		token.ReplacedByToken,
		token.ID,
	)
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to update refresh token", op))
		return apperrors.InternalServerError("token.update_error", err, nil)
	}

	return nil
}

// RevokeAllUserTokens отзывает все активные токены пользователя
func (r *refreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, ipAddress string) error {
	const op = "RefreshTokenRepository.RevokeAllUserTokens"
	if userID == uuid.Nil {
		return apperrors.BadRequestError("user_id.empty", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".RevokeAllUserTokens",
		QueryRaw: `
			UPDATE refresh_tokens
			SET revoked = true, revoked_at = $1, revoked_by_ip = $2
			WHERE user_id = $3 AND revoked = false AND expires_at > $4
		`,
	}

	_, err := r.db.ExecContext(ctx, q, time.Now(), ipAddress, userID, time.Now())
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to revoke all user tokens", op))
		return apperrors.InternalServerError("token.revoke_error", err, nil)
	}

	return nil
}

// DeleteExpired удаляет все истекшие токены
func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	const op = "RefreshTokenRepository.DeleteExpired"

	q := db.Query{
		Name: r.name + ".DeleteExpired",
		QueryRaw: `
			DELETE FROM refresh_tokens
			WHERE expires_at < $1
		`,
	}

	_, err := r.db.ExecContext(ctx, q, time.Now().Add(-24*time.Hour))
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to delete expired tokens", op))
		return apperrors.InternalServerError("token.delete_error", err, nil)
	}

	return nil
}

// CountActiveByUserID возвращает количество активных токенов пользователя
func (r *refreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	const op = "RefreshTokenRepository.CountActiveByUserID"
	if userID == uuid.Nil {
		return 0, apperrors.BadRequestError("user_id.empty", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".CountActiveByUserID",
		QueryRaw: `
			SELECT COUNT(*)
			FROM refresh_tokens
			WHERE user_id = $1 AND revoked = false AND expires_at > $2
		`,
	}

	row := r.db.QueryRowContext(ctx, q, userID, time.Now())
	var count int

	err := row.Scan(&count)
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to count active refresh tokens", op))
		return 0, apperrors.InternalServerError("token.count_error", err, nil)
	}

	return count, nil
}

// RevokeOldestIfLimitExceeded отзывает самые старые токены, если превышен лимит
func (r *refreshTokenRepository) RevokeOldestIfLimitExceeded(ctx context.Context, userID uuid.UUID, limit int, ipAddress string) error {
	const op = "RefreshTokenRepository.RevokeOldestIfLimitExceeded"
	if userID == uuid.Nil {
		return apperrors.BadRequestError("user_id.empty", nil, nil)
	}

	// Сначала получаем количество активных токенов
	count, err := r.CountActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Если лимит не превышен, ничего не делаем
	if count <= limit {
		return nil
	}

	// Иначе отзываем самые старые токены
	q := db.Query{
		Name: r.name + ".RevokeOldestIfLimitExceeded",
		QueryRaw: `
			UPDATE refresh_tokens
			SET revoked = true, revoked_at = $1, revoked_by_ip = $2
			WHERE id IN (
				SELECT id
				FROM refresh_tokens
				WHERE user_id = $3 AND revoked = false AND expires_at > $4
				ORDER BY created_at ASC
				LIMIT $5
			)
		`,
	}

	// Отзываем токены, которые превышают лимит
	tokensToRevoke := count - limit
	_, err = r.db.ExecContext(ctx, q, time.Now(), ipAddress, userID, time.Now(), tokensToRevoke)
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to revoke oldest tokens", op))
		return apperrors.InternalServerError("token.revoke_error", err, nil)
	}

	return nil
}

// GetByDeviceIdentifier находит токен по идентификатору устройства
func (r *refreshTokenRepository) GetByDeviceIdentifier(ctx context.Context, userID uuid.UUID, deviceID string) (*model.RefreshToken, error) {
	const op = "RefreshTokenRepository.GetByDeviceIdentifier"
	if userID == uuid.Nil {
		return nil, apperrors.BadRequestError("user_id.empty", nil, nil)
	}
	if deviceID == "" {
		return nil, apperrors.BadRequestError("device_id.empty", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".GetByDeviceIdentifier",
		QueryRaw: `
			SELECT id, user_id, token, expires_at, revoked, created_at, created_by_ip,
			       revoked_at, revoked_by_ip, replaced_by_token, device_identifier
			FROM refresh_tokens
			WHERE user_id = $1 AND device_identifier = $2 AND revoked = false AND expires_at > $3
			ORDER BY created_at DESC
			LIMIT 1
		`,
	}

	row := r.db.QueryRowContext(ctx, q, userID, deviceID, time.Now())
	var token model.RefreshToken

	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.Revoked,
		&token.CreatedAt,
		&token.CreatedByIP,
		&token.RevokedAt,
		&token.RevokedByIP,
		&token.ReplacedByToken,
		&token.DeviceIdentifier,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to get token by device identifier", op))
		return nil, apperrors.InternalServerError("token.get_error", err, nil)
	}

	return &token, nil
}

// RevokeByDeviceIdentifier отзывает токен для конкретного устройства
func (r *refreshTokenRepository) RevokeByDeviceIdentifier(ctx context.Context, userID uuid.UUID, deviceID string, ipAddress string) error {
	const op = "RefreshTokenRepository.RevokeByDeviceIdentifier"
	if userID == uuid.Nil {
		return apperrors.BadRequestError("user_id.empty", nil, nil)
	}
	if deviceID == "" {
		return apperrors.BadRequestError("device_id.empty", nil, nil)
	}

	q := db.Query{
		Name: r.name + ".RevokeByDeviceIdentifier",
		QueryRaw: `
			UPDATE refresh_tokens
			SET revoked = true, revoked_at = $1, revoked_by_ip = $2
			WHERE user_id = $3 AND device_identifier = $4 AND revoked = false AND expires_at > $5
		`,
	}

	_, err := r.db.ExecContext(ctx, q, time.Now(), ipAddress, userID, deviceID, time.Now())
	if err != nil {
		r.sp.Logger().WithError(err).Error(fmt.Sprintf("%s: unable to revoke token by device identifier", op))
		return apperrors.InternalServerError("token.revoke_error", err, nil)
	}

	return nil
}
