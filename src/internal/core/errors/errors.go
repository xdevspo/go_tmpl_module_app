package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

var (
	// Common domain errors
	ErrNotFound           = errors.New("errors.not_found")
	ErrInvalidInput       = errors.New("errors.invalid_input")
	ErrUnauthorized       = errors.New("errors.unauthorized")
	ErrForbidden          = errors.New("errors.forbidden")
	ErrConflict           = errors.New("errors.conflict")
	ErrValidation         = errors.New("errors.validation")
	ErrInternalServer     = errors.New("errors.internal")
	ErrServiceUnavailable = errors.New("errors.service_unavailable")
)

type AppError struct {
	Status  int         `json:"-"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error       `json:"-"`
	Stack   string      `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(status int, code, message string, err error, details interface{}) *AppError {
	appErr := &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
		Details: details,
	}

	var sb strings.Builder
	for i := 1; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, fn.Name()))
	}
	appErr.Stack = sb.String()
	return appErr
}

// Error constructors
func NotFoundError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusNotFound, "NOT_FOUND", key, err, details)
}

func BadRequestError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusBadRequest, "BAD_REQUEST", key, err, details)
}

func ValidationError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusBadRequest, "VALIDATION_ERROR", key, err, details)
}

func UnauthorizedError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", key, err, details)
}

func ForbiddenError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusForbidden, "FORBIDDEN", key, err, details)
}

func ConflictError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusConflict, "CONFLICT", key, err, details)
}

func InternalServerError(key string, err error, details any) *AppError {
	return NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", key, err, details)
}

// Error type checkers
func IsNotFoundError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "NOT_FOUND"
}

func IsValidationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "VALIDATION_ERROR"
}

func IsUnauthorizedError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "UNAUTHORIZED"
}

func IsConflictError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "CONFLICT"
}
