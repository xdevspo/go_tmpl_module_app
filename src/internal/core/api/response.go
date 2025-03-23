package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/i18n"
)

// SuccessResponse отправляет успешный ответ с заданным статусом и данными
func SuccessResponse(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// CreatedResponse отправляет ответ о создании ресурса с его ID и дополнительными данными
func CreatedResponse(c *gin.Context, id string, data any) {
	response := gin.H{
		"id": id,
	}

	if data != nil {
		response["data"] = data
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": response,
	})
}

// ActionSuccessResponse отправляет ответ об успешном выполнении действия
// message - ключ для перевода сообщения (например, "user.password_changed")
func ActionSuccessResponse(c *gin.Context, message string, details any) {
	translator := i18n.GetInstance()
	translatedMessage := translator.T(message)

	response := gin.H{
		"result":  "success",
		"message": translatedMessage,
	}

	if details != nil {
		response["details"] = details
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// ActionSuccessResponseRaw отправляет ответ об успешном выполнении действия
// message - готовое сообщение без перевода
func ActionSuccessResponseRaw(c *gin.Context, message string, details any) {
	response := gin.H{
		"result":  "success",
		"message": message,
	}

	if details != nil {
		response["details"] = details
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// NoContentResponse отправляет пустой ответ со статусом 204
func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
