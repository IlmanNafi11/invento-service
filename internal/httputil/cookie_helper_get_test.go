package httputil_test

import (
	"net/http"
	"testing"

	"invento-service/config"
	"invento-service/internal/httputil"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetRefreshTokenFromCookie(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{App: config.AppConfig{Env: "development"}}
	cookieHelper := httputil.NewCookieHelper(cfg)
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		result := cookieHelper.GetRefreshTokenFromCookie(c)
		return c.SendString(result)
	})

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "Token exists",
			token:    "valid-token-123",
			expected: "valid-token-123",
		},
		{
			name:     "Empty token",
			token:    "",
			expected: "",
		},
		{
			name:     "Long token string",
			token:    "very-long-refresh-token-with-many-characters-123456789",
			expected: "very-long-refresh-token-with-many-characters-123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", http.NoBody)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: httputil.RefreshTokenCookieName, Value: tt.token})
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)

			body := make([]byte, len(tt.expected))
			resp.Body.Read(body)
			assert.Equal(t, tt.expected, string(body))
		})
	}
}

func TestCookieHelper_GetAccessTokenFromCookie(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{App: config.AppConfig{Env: "development"}}
	cookieHelper := httputil.NewCookieHelper(cfg)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(cookieHelper.GetAccessTokenFromCookie(c))
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.AddCookie(&http.Cookie{Name: httputil.AccessTokenCookieName, Value: "cookie-access-token"})

	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := make([]byte, len("cookie-access-token"))
	_, _ = resp.Body.Read(body)
	assert.Equal(t, "cookie-access-token", string(body))
}

func TestCookieHelper_ClearAllAuthCookies(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{App: config.AppConfig{Env: "development"}}
	cookieHelper := httputil.NewCookieHelper(cfg)
	app := fiber.New()

	app.Post("/clear", func(c *fiber.Ctx) error {
		cookieHelper.ClearAllAuthCookies(c)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("POST", "/clear", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	seen := map[string]*http.Cookie{}
	for _, c := range resp.Cookies() {
		if c.Name == httputil.AccessTokenCookieName || c.Name == httputil.RefreshTokenCookieName {
			seen[c.Name] = c
		}
	}

	assert.NotNil(t, seen[httputil.AccessTokenCookieName])
	assert.Equal(t, "", seen[httputil.AccessTokenCookieName].Value)
	assert.Equal(t, 0, seen[httputil.AccessTokenCookieName].MaxAge)

	assert.NotNil(t, seen[httputil.RefreshTokenCookieName])
	assert.Equal(t, "", seen[httputil.RefreshTokenCookieName].Value)
	assert.Equal(t, 0, seen[httputil.RefreshTokenCookieName].MaxAge)
}

func TestCookieHelper_SecurityProperties(t *testing.T) {
	t.Parallel()
	t.Run("Production cookie has secure properties", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "production",
			},
		}

		cookieHelper := httputil.NewCookieHelper(cfg)
		app := fiber.New()

		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "prod-token")
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", http.NoBody)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		cookies := resp.Cookies()
		var refreshToken *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshToken = c
				break
			}
		}

		assert.NotNil(t, refreshToken)
		assert.Equal(t, "prod-token", refreshToken.Value)
		assert.Equal(t, true, refreshToken.Secure)
		assert.Equal(t, true, refreshToken.HttpOnly)
	})

	t.Run("Development cookie properties", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
		}

		cookieHelper := httputil.NewCookieHelper(cfg)
		app := fiber.New()

		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "dev-token")
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", http.NoBody)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		cookies := resp.Cookies()
		var refreshToken *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshToken = c
				break
			}
		}

		assert.NotNil(t, refreshToken)
		assert.Equal(t, "dev-token", refreshToken.Value)
		assert.Equal(t, false, refreshToken.Secure)
		assert.Equal(t, true, refreshToken.HttpOnly)
	})
}

func TestCookieHelper_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Special characters in token", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
		}

		cookieHelper := httputil.NewCookieHelper(cfg)
		app := fiber.New()

		specialToken := "token.with.dots-and_underscores"
		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, specialToken)
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", http.NoBody)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		cookies := resp.Cookies()
		var refreshToken *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshToken = c
				break
			}
		}

		assert.NotNil(t, refreshToken)
		assert.Equal(t, specialToken, refreshToken.Value)
	})

	t.Run("URL-encoded characters in token", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
		}

		cookieHelper := httputil.NewCookieHelper(cfg)
		app := fiber.New()

		// Use URL-safe characters instead of raw unicode
		specialToken := "token-with-dashes_and_underscores-123"
		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, specialToken)
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", http.NoBody)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		cookies := resp.Cookies()
		var refreshToken *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshToken = c
				break
			}
		}

		assert.NotNil(t, refreshToken)
		assert.Equal(t, specialToken, refreshToken.Value)
	})

	t.Run("Very long expiration", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
		}

		cookieHelper := httputil.NewCookieHelper(cfg)
		app := fiber.New()

		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "long-expiry-token")
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", http.NoBody)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		cookies := resp.Cookies()
		var refreshToken *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshToken = c
				break
			}
		}

		assert.NotNil(t, refreshToken)
		assert.Equal(t, "long-expiry-token", refreshToken.Value)
	})
}

func TestCookieHelper_Integration(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
	}

	cookieHelper := httputil.NewCookieHelper(cfg)
	app := fiber.New()

	t.Run("Complete cookie lifecycle", func(t *testing.T) {
		// Test setting a cookie
		app.Post("/set", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "test-token")
			return c.SendString("ok")
		})

		req1, _ := http.NewRequest("POST", "/set", http.NoBody)
		resp1, _ := app.Test(req1)

		// Check cookie was set in response
		cookies1 := resp1.Cookies()
		var cookie1 *http.Cookie
		for _, c := range cookies1 {
			if c.Name == "refresh_token" {
				cookie1 = c
				break
			}
		}
		assert.NotNil(t, cookie1)
		assert.Equal(t, "test-token", cookie1.Value)

		// Test clearing a cookie
		app.Post("/clear", func(c *fiber.Ctx) error {
			cookieHelper.ClearRefreshTokenCookie(c)
			return c.SendString("ok")
		})

		req2, _ := http.NewRequest("POST", "/clear", http.NoBody)
		resp2, _ := app.Test(req2)

		// Check cookie was cleared
		cookies2 := resp2.Cookies()
		var cookie2 *http.Cookie
		for _, c := range cookies2 {
			if c.Name == "refresh_token" {
				cookie2 = c
				break
			}
		}
		// Cleared cookie should have empty value and MaxAge = -1
		assert.NotNil(t, cookie2)
		assert.Equal(t, "", cookie2.Value)
		assert.Equal(t, 0, cookie2.MaxAge)
	})

	t.Run("Multiple requests independence", func(t *testing.T) {
		app.Post("/token1", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "token-1")
			return c.SendString("ok")
		})

		app.Post("/token2", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "token-2")
			return c.SendString("ok")
		})

		// Each request is independent
		req1, _ := http.NewRequest("POST", "/token1", http.NoBody)
		resp1, _ := app.Test(req1)
		cookies1 := resp1.Cookies()
		var cookie1 *http.Cookie
		for _, c := range cookies1 {
			if c.Name == "refresh_token" {
				cookie1 = c
				break
			}
		}
		assert.Equal(t, "token-1", cookie1.Value)

		req2, _ := http.NewRequest("POST", "/token2", http.NoBody)
		resp2, _ := app.Test(req2)
		cookies2 := resp2.Cookies()
		var cookie2 *http.Cookie
		for _, c := range cookies2 {
			if c.Name == "refresh_token" {
				cookie2 = c
				break
			}
		}
		assert.Equal(t, "token-2", cookie2.Value)
	})
}
