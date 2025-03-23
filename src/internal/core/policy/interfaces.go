package policy

import (
	"context"

	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

type Policy interface {
	Check(ctx context.Context, user *model.User, resource string, action string) bool
}
