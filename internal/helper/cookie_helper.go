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

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Strict",
		MaxAge:   ch.config.JWT.RefreshTokenExpireHours * 3600,
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
