package errors

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/i18n"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/logger"
)

func ResponseWithError(c *gin.Context, err error) {
	var appErr *AppError
	var validErrs validator.ValidationErrors

	if errors.As(err, &appErr) {
		// Already an AppError
	} else if errors.As(err, &validErrs) {
		// Validation error
		appErr = ValidationError("errors.validation", err, parseValidationErrors(err))
	} else {
		// Unknown error
		appErr = InternalServerError("errors.internal", err, nil)
	}

	loggerObj, exists := c.Get("logger")
	if exists {
		if loggerInstance, ok := loggerObj.(logger.Logger); ok {
			if appErr.Status >= 500 {
				loggerInstance.WithFields(logrus.Fields{
					"error": appErr.Err,
					"stack": appErr.Stack,
				}).Error(appErr.Message)
			} else {
				loggerInstance.WithField("error", appErr.Err).Warn(appErr.Message)
			}
		}
	}

	translator := i18n.GetInstance()
	displayMessage := translator.T(appErr.Message)

	c.JSON(appErr.Status, gin.H{
		"error": gin.H{
			"code":    appErr.Code,
			"message": displayMessage,
			"details": appErr.Details,
		},
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = c.Query("lang")
		}
		if lang != "" {
			i18n.GetInstance().SetLanguage(lang)
		}

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			ResponseWithError(c, err)
		}
	}
}
