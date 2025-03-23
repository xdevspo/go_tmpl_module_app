# Система политик доступа (Access Policy)

## Общее описание

Система политик доступа представляет собой реализацию паттерна Policy для контроля доступа к ресурсам API. Она позволяет декларативно определять правила доступа к различным ресурсам системы на основе ролей и разрешений пользователя.

## Компоненты системы

### 1. Интерфейс Policy

Центральным элементом системы является интерфейс `Policy`, который определяет контракт для всех политик доступа:

```go
type Policy interface {
    Check(ctx context.Context, user *model.User, resource string, action string) bool
}
```

Этот интерфейс принимает:
- `ctx` - контекст запроса
- `user` - проверяемый пользователь
- `resource` - ресурс, к которому запрашивается доступ
- `action` - действие, которое пользователь пытается выполнить
- Возвращает `bool` - разрешен ли доступ

### 2. Фабрика политик (PolicyFactory)

Фабрика политик служит для:
- Регистрации всех политик в системе
- Получения нужной политики по имени ресурса
- Централизованного управления политиками

```go
type PolicyFactory struct {
    policies map[string]Policy
}
```

Основные методы:
- `NewPolicyFactory()` - создает фабрику и регистрирует базовые политики
- `ForResource(resource string) (Policy, error)` - возвращает политику для указанного ресурса
- `RegisterPolicy(resource string, policy Policy)` - регистрирует новую политику

### 3. Middleware для проверки политик

`PolicyMiddleware` применяет политики в контексте HTTP-запросов:

```go
func (m *PolicyMiddleware) RequirePermission(resource string, action string) gin.HandlerFunc {
    // Проверяет наличие пользователя в контексте
    // Получает соответствующую политику
    // Проверяет права доступа
    // Прерывает запрос если доступ запрещен
}
```

## Архитектура системы политик

### 1. Модульная структура

Каждый модуль системы может определять свои политики доступа:

```
internal/
  ├── core/
  │   └── policy/         # Ядро системы политик
  │       ├── interfaces.go  # Определение интерфейса Policy
  │       └── factory.go     # Реализация PolicyFactory
  │
  └── module/
      ├── user/           # Модуль пользователей
      │   └── policy/     # Политики модуля пользователей
      │       └── user_policy.go
      │
      └── another-module/ # Другой модуль
          └── policy/     # Политики другого модуля
```

### 2. Регистрация политик

При инициализации системы каждый модуль регистрирует свои политики:

```go
// В router.go или другом месте инициализации
policyFactory := corepolicy.NewPolicyFactory()

// Регистрация политик разных модулей
userPolicy.RegisterInFactory(policyFactory)
anotherModulePolicy.RegisterInFactory(policyFactory)
```

### 3. Применение политик в маршрутах

```go
// В handler.go каждого модуля
func (h *SomeHandler) Register(group *gin.RouterGroup, policyMiddleware *middleware.PolicyMiddleware) {
    // Публичные маршруты без проверки политик
    group.GET("/public", h.PublicEndpoint)
    
    // Защищенные маршруты с проверкой политик
    group.GET("/protected", 
        policyMiddleware.RequirePermission("resource-name", "view"), 
        h.ProtectedEndpoint)
}
```

## Создание новой политики

Для добавления новой политики нужно:

1. Создать файл политики в соответствующем модуле:

```go
// internal/module/mymodule/policy/my_policy.go
package policy

import (
    "context"
    corepolicy "github.com/xdevspo/go_tmpl_module_app/internal/core/policy"
    "github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

// Имя ресурса
const ResourceName = "my-resource"

// Реализация политики
type MyPolicy struct{}

func NewMyPolicy() *MyPolicy {
    return &MyPolicy{}
}

// Проверка прав доступа
func (p *MyPolicy) Check(ctx context.Context, user *model.User, resource string, action string) bool {
    // Логика проверки прав
    switch action {
    case "view":
        return user.HasPermission("my-resource:view")
    case "edit":
        return user.HasPermission("my-resource:edit")
    }
    return false
}

// Регистрация политики в фабрике
func RegisterInFactory(factory *corepolicy.PolicyFactory) {
    factory.RegisterPolicy(ResourceName, NewMyPolicy())
}
```

2. Добавить вызов регистрации в `router.go` или другое место инициализации:

```go
// В registerModulePolicies или аналогичном месте
myModulePolicy.RegisterInFactory(policyFactory)
```

3. Использовать в маршрутах:

```go
group.POST("/my-resource", 
    policyMiddleware.RequirePermission("my-resource", "edit"), 
    h.CreateMyResource)
```

## Преимущества системы

1. **Декларативность** - четкое описание правил доступа
2. **Модульность** - политики инкапсулированы в соответствующих модулях
3. **Расширяемость** - легко добавлять новые политики и ресурсы
4. **Единообразие** - унифицированный подход к проверке доступа
5. **Безопасность** - централизованная логика проверки доступа
6. **Тестируемость** - легко тестировать политики в изоляции

## Рекомендуемые практики

1. Используйте константы для имен ресурсов и действий
2. Размещайте политики внутри соответствующих модулей
3. Используйте четкие имена разрешений (например, `users:create`)
4. При сложной логике добавляйте документацию к методу `Check()`
5. Не мешайте бизнес-логику с логикой проверки доступа 
