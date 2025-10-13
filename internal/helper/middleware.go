package helper

import (
	"log"
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

func TusProtocolMiddleware(tusVersion string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()

		if method == "OPTIONS" {
			c.Set("Tus-Resumable", tusVersion)
			c.Set("Tus-Version", tusVersion)
			c.Set("Tus-Extension", "creation,termination")
			c.Set("Tus-Max-Size", "524288000")
			return c.SendStatus(fiber.StatusNoContent)
		}

		if method == "PATCH" {
			// Log all headers for PATCH request
			log.Printf("=== TUS MIDDLEWARE DEBUG - PATCH Request ===")
			log.Printf("Path: %s", c.Path())
			log.Printf("Headers:")
			c.Request().Header.VisitAll(func(key, value []byte) {
				log.Printf("  %s: %s", string(key), string(value))
			})
			log.Printf("===========================================")
		}

		tusResumable := c.Get("Tus-Resumable")
		if method != "GET" && method != "POST" && tusResumable == "" {
			log.Printf("TUS MIDDLEWARE: Missing Tus-Resumable header on %s request", method)
			c.Set("Tus-Resumable", tusVersion)
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}

		if tusResumable != "" && tusResumable != tusVersion {
			log.Printf("TUS MIDDLEWARE: Invalid Tus-Resumable version: got %s, expected %s", tusResumable, tusVersion)
			c.Set("Tus-Resumable", tusVersion)
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}

		return c.Next()
	}
}
