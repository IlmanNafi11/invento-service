package helper

import (
	"fiber-boiler-plate/config"

	"github.com/gofiber/fiber/v2"
)

type CookieHelper struct {
	config *config.Config
}

func NewCookieHelper(cfg *config.Config) *CookieHelper {
	return &CookieHelper{
		config: cfg,
	}
}

func (ch *CookieHelper) SetRefreshTokenCookie(c *fiber.Ctx, token string) {
	secure := ch.config.App.Env == "production"
	// Default 7 days (604800 seconds) for refresh token cookie
	// Supabase refresh tokens are typically valid for 7 days
	maxAge := 604800

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Strict",
		MaxAge:   maxAge,
		Path:     "/",
	})
}

func (ch *CookieHelper) ClearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   ch.config.App.Env == "production",
		SameSite: "Strict",
		MaxAge:   -1,
		Path:     "/",
	})
}

func GetRefreshTokenFromCookie(c *fiber.Ctx) string {
	return c.Cookies("refresh_token")
}
