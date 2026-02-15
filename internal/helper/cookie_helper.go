package helper

import (
	"invento-service/config"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
	AccessTokenPath        = "/"
	RefreshTokenPath       = "/api/v1/auth"
)

type CookieHelper struct {
	config *config.Config
}

func NewCookieHelper(cfg *config.Config) *CookieHelper {
	return &CookieHelper{
		config: cfg,
	}
}

// SetAccessTokenCookie stores the access token in an HttpOnly cookie.
// expiresIn is the token lifetime in seconds (from Supabase).
func (ch *CookieHelper) SetAccessTokenCookie(c *fiber.Ctx, token string, expiresIn int) {
	c.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookieName,
		Value:    token,
		Expires:  time.Now().Add(time.Duration(expiresIn) * time.Second),
		HTTPOnly: true,
		Secure:   ch.config.App.Env == "production",
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     AccessTokenPath,
	})
}

// SetRefreshTokenCookie stores the refresh token in an HttpOnly cookie.
// Path is restricted to /api/v1/auth to prevent unnecessary transmission.
func (ch *CookieHelper) SetRefreshTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   ch.config.App.Env == "production",
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     RefreshTokenPath,
	})
}

// GetAccessTokenFromCookie reads the access token from the cookie.
func (ch *CookieHelper) GetAccessTokenFromCookie(c *fiber.Ctx) string {
	return c.Cookies(AccessTokenCookieName)
}

// GetRefreshTokenFromCookie reads the refresh token from the cookie.
func (ch *CookieHelper) GetRefreshTokenFromCookie(c *fiber.Ctx) string {
	return c.Cookies(RefreshTokenCookieName)
}

// ClearRefreshTokenCookie removes the refresh token cookie.
func (ch *CookieHelper) ClearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   ch.config.App.Env == "production",
		SameSite: fiber.CookieSameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     RefreshTokenPath,
	})
}

// ClearAllAuthCookies removes both access and refresh token cookies.
func (ch *CookieHelper) ClearAllAuthCookies(c *fiber.Ctx) {
	ch.ClearRefreshTokenCookie(c)
	c.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   ch.config.App.Env == "production",
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     AccessTokenPath,
	})
}
