package policy

import (
	"context"

	corepolicy "github.com/xdevspo/go_tmpl_module_app/internal/core/policy"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

// Название ресурса, используемое в маршрутах при проверке доступа
const ResourceName = "user"

type UserPolicy struct{}

func NewUserPolicy() *UserPolicy {
	return &UserPolicy{}
}

func (p *UserPolicy) Check(ctx context.Context, user *model.User, resource string, action string) bool {
	if user.HasAnyPermission("full", "users:full") {
		return true
	}

	switch action {
	case "create":
		return user.HasPermission("users:create")
	case "view":
		return user.HasPermission("users:view")
	case "update":
		return user.HasPermission("users:update")
	case "delete":
		return user.HasPermission("users:delete")
	case "assign-role":
		return user.HasPermission("users:assign-role")
	case "revoke-role":
		return user.HasPermission("users:revoke-role")
	case "view-roles":
		return user.HasPermission("users:view-roles")
	case "assign-permission":
		return user.HasPermission("users:assign-permission")
	case "revoke-permission":
		return user.HasAnyPermission("users:revoke-permission")
	case "view-permissions":
		return user.HasPermission("users:view-permissions")
	default:
		return false
	}
}

// RegisterInFactory регистрирует политику пользователей в центральной фабрике политик
func RegisterInFactory(factory *corepolicy.PolicyFactory) {
	factory.RegisterPolicy(ResourceName, NewUserPolicy())
}
