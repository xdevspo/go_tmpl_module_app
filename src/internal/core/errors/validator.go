package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entrans "github.com/go-playground/validator/v10/translations/en"
	rutrans "github.com/go-playground/validator/v10/translations/ru"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/provider"
)

type ValidationFieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

func InitValidator(sp provider.ServiceProvider) error {
	enLocale := en.New()
	uni = ut.New(enLocale)

	ruLocale := ru.New()
	if err := uni.AddTranslator(ruLocale, true); err != nil {
		return fmt.Errorf("failed to add Russian translator: %w", err)
	}

	lang := sp.AppConfig().Lang()
	var found bool
	trans, found = uni.GetTranslator(lang)
	if !found {
		trans, _ = uni.GetTranslator("en")
	}

	validate = validator.New()

	var err error
	switch lang {
	case "ru":
		err = rutrans.RegisterDefaultTranslations(validate, trans)
	default:
		err = entrans.RegisterDefaultTranslations(validate, trans)
	}
	if err != nil {
		return fmt.Errorf("failed to register translations: %w", err)
	}

	return nil
}

func ValidateRequest(obj interface{}) error {
	if err := validate.Struct(obj); err != nil {
		return ValidationError("errors.validation", err, parseValidationErrors(err))
	}
	return nil
}

func parseValidationErrors(err error) []ValidationFieldError {
	var validationErrors []ValidationFieldError
	var validErrs validator.ValidationErrors

	if err == nil {
		return validationErrors
	}

	if errors.As(err, &validErrs) {
		translatedErrs := validErrs.Translate(trans)

		for _, fieldErr := range validErrs {
			message := translatedErrs[fieldErr.Namespace()]
			validationField := ValidationFieldError{
				Field:   fieldErr.Field(),
				Tag:     fieldErr.Tag(),
				Message: message,
				Param:   fieldErr.Param(),
			}

			if !isConfidentialField(fieldErr.Field()) {
				validationField.Value = fieldErr.Value()
			}

			validationErrors = append(validationErrors, validationField)
		}
	}

	return validationErrors
}

func isConfidentialField(fieldName string) bool {
	lowercaseName := strings.ToLower(fieldName)
	confidentialFields := []string{"password", "пароль", "secret", "token", "key", "pin"}

	for _, field := range confidentialFields {
		if strings.Contains(lowercaseName, field) {
			return true
		}
	}

	return false
}
