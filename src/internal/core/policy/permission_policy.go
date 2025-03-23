package policy

import (
	"context"

	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

type PermissionPolicy struct{}

func NewPermissionPolicy() *PermissionPolicy {
	return &PermissionPolicy{}
}

func (p *PermissionPolicy) Check(ctx context.Context, user *model.User, resource string, action string) bool {
	if user.HasAnyPermission("full", "permissions:full") {
		return true
	}

	switch action {
	case "create":
		return user.HasPermission("permissions:create")
	case "view":
		return user.HasPermission("permissions:view")
	case "delete":
		return user.HasPermission("permissions:delete")
	case "assign":
		return user.HasPermission("permissions:assign")
	case "revoke":
		return user.HasPermission("permissions:revoke")
	}

	return false
}
