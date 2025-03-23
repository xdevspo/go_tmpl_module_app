# API Response Helpers

Этот пакет содержит вспомогательные функции для стандартизации API ответов.

## Типы ответов

1. **SuccessResponse** - для отправки успешного ответа с данными:
   ```json
   {
     "data": { ... } // любые данные
   }
   ```

2. **CreatedResponse** - для отправки ответа о создании ресурса:
   ```json
   {
     "data": {
       "id": "123",
       "data": { ... } // опционально
     }
   }
   ```

3. **ActionSuccessResponse** - для отправки ответа об успешном выполнении действия с переводом сообщения:
   ```json
   {
     "data": {
       "result": "success",
       "message": "операция выполнена", // переведенное сообщение
       "details": { ... } // опционально
     }
   }
   ```

4. **ActionSuccessResponseRaw** - для отправки ответа об успешном выполнении действия без перевода:
   ```json
   {
     "data": {
       "result": "success",
       "message": "операция выполнена", // исходное сообщение без перевода
       "details": { ... } // опционально
     }
   }
   ```

5. **NoContentResponse** - для отправки пустого ответа (статус 204)

## Поддержка интернационализации (i18n)

Функция `ActionSuccessResponse` автоматически переводит сообщение используя пакет i18n:

```go
// Передайте ключ для перевода вместо готового сообщения
api.ActionSuccessResponse(c, "response.user.role_assigned", nil)
```

В файлах переводов (например, `ru.json`) добавьте соответствующий ключ:

```json
{
  "response.user.role_assigned": "Роль успешно назначена"
}
```

Если перевод не нужен, используйте `ActionSuccessResponseRaw`:

```go
// Используйте готовое сообщение без перевода
api.ActionSuccessResponseRaw(c, "Роль успешно назначена", nil)
```

## Примеры использования

```go
// Возврат сущности
api.SuccessResponse(c, user)

// Возврат ответа о создании
api.CreatedResponse(c, user.ID.String(), nil)

// Возврат ответа об успешном действии с переводом
api.ActionSuccessResponse(c, "response.user.role_assigned", nil)

// Возврат ответа об успешном действии без перевода
api.ActionSuccessResponseRaw(c, "Роль успешно назначена", nil)

// Возврат пустого ответа
api.NoContentResponse(c)
```

## Преимущества

- **Единообразие** - все API ответы имеют одинаковую структуру
- **Предсказуемость** - клиенты могут рассчитывать на определенный формат ответов
- **Локализация** - автоматический перевод сообщений с помощью i18n
- **Читаемость** - структура ответов понятна и логична
- **Расширяемость** - структура ответов может быть расширена дополнительными полями 

## Пример полного обработчика с поддержкой i18n

```go
// Пример хендлера назначения роли пользователю
func (h *UserHandler) AssignRoleHandler(c *gin.Context) {
    var req struct {
        UserID uuid.UUID `json:"user_id" binding:"required"`
        RoleID int       `json:"role_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        // Передаем ошибку валидации в общий обработчик ошибок
        apperrors.ResponseWithError(c, err)
        return
    }

    if err := h.userService.AssignRole(c.Request.Context(), req.UserID, req.RoleID); err != nil {
        // Передаем ошибку сервиса в общий обработчик ошибок
        apperrors.ResponseWithError(c, err)
        return
    }

    // Используем ключ для перевода вместо готового сообщения
    api.ActionSuccessResponse(c, "response.user.role_assigned", nil)
}
```

В файле переводов ru.json:
```json
{
  "response.user.role_assigned": "Роль успешно назначена"
}
```

В файле переводов en.json:
```json
{
  "response.user.role_assigned": "Role successfully assigned"
}
```

Теперь, в зависимости от выбранного языка, клиент получит соответствующий ответ:

Для русского языка:
```json
{
  "data": {
    "result": "success",
    "message": "Роль успешно назначена"
  }
}
```

Для английского языка:
```json
{
  "data": {
    "result": "success",
    "message": "Role successfully assigned"
  }
}
``` 
