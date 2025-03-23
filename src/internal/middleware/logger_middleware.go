package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
)

// Define a custom type for context keys to avoid collisions
type contextKey string

// RequestIDKey is the key used to store request ID in context
const RequestIDKey contextKey = "request_id"

// RequestIDGinKey is the string key used for Gin context
const RequestIDGinKey = "request_id"

// RequestLoggerWithLogger creates a middleware that logs information about incoming HTTP requests using a provided logger
func RequestLoggerWithLogger(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique request ID for tracing
		requestID := uuid.New().String()

		// Store the request ID in both Gin context and HTTP headers
		c.Set(RequestIDGinKey, requestID)
		c.Header("X-Request-ID", requestID)

		// Record start time for calculating request duration
		startTime := time.Now()

		// Create fields for structured logging
		fields := logrus.Fields{
			"client_ip":   c.ClientIP(),
			"request_id":  requestID,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"query":       c.Request.URL.RawQuery,
			"user_agent":  c.Request.UserAgent(),
			"status_code": 0, // Will be updated after request processing
		}

		// Process the request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)
		fields["status_code"] = c.Writer.Status()
		fields["duration"] = duration.String()
		fields["response_size"] = c.Writer.Size()

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logger.WithFields(fields).Error("Request failed")
		} else if c.Writer.Status() >= 400 {
			logger.WithFields(fields).Warn("Request completed with client error")
		} else {
			logger.WithFields(fields).Info("Request completed successfully")
		}
	}
}

// WithRequestID attaches a requestID to the given context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the requestID from the given context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
