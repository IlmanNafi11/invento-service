package middleware

import (
	"invento-service/internal/domain"
	"invento-service/internal/httputil"
	"invento-service/internal/usecase/repo"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SupabaseAuthMiddleware validates Supabase JWT tokens and extracts user info
func SupabaseAuthMiddleware(authService domain.AuthService, userRepo repo.UserRepository, cookieHelper *httputil.CookieHelper) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := ""
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return httputil.SendErrorResponse(c, fiber.StatusUnauthorized, "Format token tidak valid", nil)
			}
			accessToken = tokenParts[1]
		}

		if accessToken == "" && cookieHelper != nil {
			accessToken = cookieHelper.GetAccessTokenFromCookie(c)
		}

		if accessToken == "" {
			return httputil.SendUnauthorizedResponse(c)
		}

		claims, err := authService.VerifyJWT(accessToken)
		if err != nil {
			return httputil.SendUnauthorizedResponse(c)
		}

		user, err := userRepo.GetByID(claims.GetUserID())
		if err != nil {
			return httputil.SendUnauthorizedResponse(c)
		}

		if !user.IsActive {
			return httputil.SendUnauthorizedResponse(c)
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
