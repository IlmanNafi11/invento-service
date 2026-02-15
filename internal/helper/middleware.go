package helper

import (
	"invento-service/internal/domain"
	"invento-service/internal/usecase/repo"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	zlog "github.com/rs/zerolog/log"
)

// CasbinPermissionChecker is an interface for checking permissions.
// This allows for easy mocking in tests.
type CasbinPermissionChecker interface {
	CheckPermission(roleName, resource, action string) (bool, error)
}

// SupabaseAuthMiddleware validates Supabase JWT tokens and extracts user info
func SupabaseAuthMiddleware(authService domain.AuthService, userRepo repo.UserRepository, cookieHelper *CookieHelper) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := ""
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return SendErrorResponse(c, fiber.StatusUnauthorized, "Format token tidak valid", nil)
			}
			accessToken = tokenParts[1]
		}

		if accessToken == "" && cookieHelper != nil {
			accessToken = cookieHelper.GetAccessTokenFromCookie(c)
		}

		if accessToken == "" {
			return SendUnauthorizedResponse(c)
		}

		claims, err := authService.VerifyJWT(accessToken)
		if err != nil {
			return SendUnauthorizedResponse(c)
		}

		user, err := userRepo.GetByID(claims.GetUserID())
		if err != nil {
			return SendUnauthorizedResponse(c)
		}

		if !user.IsActive {
			return SendUnauthorizedResponse(c)
		}

		c.Locals("user_id", user.ID)
		c.Locals("user_email", user.Email)

		roleName := ""
		if user.Role != nil {
			roleName = user.Role.NamaRole
		}
		c.Locals("user_role", roleName)
		c.Locals("access_token", accessToken)
		c.Locals("claims", claims)

		return c.Next()
	}
}

func RBACMiddleware(casbinEnforcer CasbinPermissionChecker, resource string, action string) fiber.Handler {
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
			zlog.Error().Err(err).Str("role", role).Str("resource", resource).Str("action", action).Msg("RBAC CheckPermission failed")
			return SendInternalServerErrorResponse(c)
		}

		if !allowed {
			return SendForbiddenResponse(c)
		}

		return c.Next()
	}
}

func TusProtocolMiddleware(tusVersion string, maxSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()

		if method == "OPTIONS" {
			c.Set("Tus-Resumable", tusVersion)
			c.Set("Tus-Version", tusVersion)
			c.Set("Tus-Extension", "creation,termination")
			c.Set("Tus-Max-Size", strconv.FormatInt(maxSize, 10))
			return c.SendStatus(fiber.StatusNoContent)
		}

		tusResumable := c.Get("Tus-Resumable")
		if method != "GET" && method != "POST" && tusResumable == "" {
			c.Set("Tus-Resumable", tusVersion)
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}

		if tusResumable != "" && tusResumable != tusVersion {
			c.Set("Tus-Resumable", tusVersion)
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}

		return c.Next()
	}
}
