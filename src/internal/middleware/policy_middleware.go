package middleware

import (
	"github.com/gin-gonic/gin"
	apperrors "github.com/xdevspo/go_tmpl_module_app/internal/core/errors"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/policy"
	"github.com/xdevspo/go_tmpl_module_app/internal/module/user/model"
)

type PolicyMiddleware struct {
	policyFactory *policy.PolicyFactory
}

func NewPolicyMiddleware(policyFactory *policy.PolicyFactory) *PolicyMiddleware {
	return &PolicyMiddleware{policyFactory: policyFactory}
}

func (m *PolicyMiddleware) RequirePermission(resource string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			apperrors.ResponseWithError(c, apperrors.UnauthorizedError("errors.unauthorized", nil, nil))
			c.Abort()
			return
		}

		currentUser, ok := user.(*model.User)
		if !ok {
			apperrors.ResponseWithError(c, apperrors.InternalServerError("errors.internal", nil, nil))
			c.Abort()
			return
		}

		policy, err := m.policyFactory.ForResource(resource)
		if err != nil {
			apperrors.ResponseWithError(c, apperrors.ForbiddenError("errors.forbidden", err, map[string]interface{}{
				"message":  "политика доступа не найдена",
				"resource": resource,
			}))
			c.Abort()
			return
		}

		if !policy.Check(c.Request.Context(), currentUser, resource, action) {
			apperrors.ResponseWithError(c, apperrors.ForbiddenError("errors.forbidden", nil, map[string]interface{}{
				"message":  "недостаточно прав для выполнения операции",
				"resource": resource,
				"action":   action,
			}))
			c.Abort()
			return
		}

		c.Next()
	}
}
