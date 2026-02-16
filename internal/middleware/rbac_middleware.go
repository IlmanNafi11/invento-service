package middleware

import (
	"invento-service/internal/httputil"

	"github.com/gofiber/fiber/v2"
	zlog "github.com/rs/zerolog/log"
)

// CasbinPermissionChecker is an interface for checking permissions.
// This allows for easy mocking in tests.
type CasbinPermissionChecker interface {
	CheckPermission(roleName, resource, action string) (bool, error)
}

func RBACMiddleware(casbinEnforcer CasbinPermissionChecker, resource string, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleVal := c.Locals(LocalsKeyUserRole)
		if roleVal == nil {
			return httputil.SendForbiddenResponse(c)
		}

		role, ok := roleVal.(string)
		if !ok || role == "" {
			return httputil.SendForbiddenResponse(c)
		}

		allowed, err := casbinEnforcer.CheckPermission(role, resource, action)
		if err != nil {
			zlog.Error().Err(err).Str("role", role).Str("resource", resource).Str("action", action).Msg("RBAC CheckPermission failed")
			return httputil.SendInternalServerErrorResponse(c)
		}

		if !allowed {
			return httputil.SendForbiddenResponse(c)
		}

		return c.Next()
	}
}
