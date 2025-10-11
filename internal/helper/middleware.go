package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTAuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return SendUnauthorizedResponse(c)
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return SendErrorResponse(c, fiber.StatusUnauthorized, "Format token tidak valid", nil)
		}

		claims, err := ValidateAccessToken(tokenParts[1], secret)
		if err != nil {
			return SendErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		return c.Next()
	}
}

func RBACMiddleware(casbinEnforcer *CasbinEnforcer, resource string, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleVal := c.Locals("user_role")
		if roleVal == nil {
			return SendForbiddenResponse(c)
		}

		role, ok := roleVal.(string)
		if !ok || role == "" {
			return SendForbiddenResponse(c)
		}

		allowed, err := casbinEnforcer.CheckPermission(role, resource, action)
		if err != nil {
			return SendInternalServerErrorResponse(c)
		}

		if !allowed {
			return SendForbiddenResponse(c)
		}

		return c.Next()
	}
}
