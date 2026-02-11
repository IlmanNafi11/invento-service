package helper_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/helper"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewCookieHelper(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 24,
		},
	}

	cookieHelper := helper.NewCookieHelper(cfg)

	assert.NotNil(t, cookieHelper)
}

func TestCookieHelper_SetRefreshTokenCookie(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		token  string
		secure bool
	}{
		{
			name:   "Production environment - secure cookie",
			env:    "production",
			token:  "valid-refresh-token-123",
			secure: true,
		},
		{
			name:   "Development environment - not secure",
			env:    "development",
			token:  "dev-refresh-token-456",
			secure: false,
		},
		{
			name:   "Staging environment - not secure",
			env:    "staging",
			token:  "staging-refresh-token-789",
			secure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{
					Env: tt.env,
				},
				JWT: config.JWTConfig{
					RefreshTokenExpireHours: 24,
				},
			}

			cookieHelper := helper.NewCookieHelper(cfg)
			app := fiber.New()

			// Create a test app and handler that sets the cookie
			app.Post("/test", func(c *fiber.Ctx) error {
				cookieHelper.SetRefreshTokenCookie(c, tt.token)
				return c.SendString("ok")
			})

			// Make a test request
			req, _ := http.NewRequest("POST", "/test", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)

			// Get cookie from response
			cookies := resp.Cookies()
			var refreshToken *http.Cookie
			for _, c := range cookies {
				if c.Name == "refresh_token" {
					refreshToken = c
					break
				}
			}

			assert.NotNil(t, refreshToken)
			assert.Equal(t, tt.token, refreshToken.Value)
			assert.Equal(t, tt.secure, refreshToken.Secure)
		})
	}
}

func TestCookieHelper_SetRefreshTokenCookie_Properties(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 48,
		},
	}

	cookieHelper := helper.NewCookieHelper(cfg)
	app := fiber.New()

	testToken := "test-refresh-token"
	app.Post("/test", func(c *fiber.Ctx) error {
		cookieHelper.SetRefreshTokenCookie(c, testToken)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
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
	assert.Equal(t, testToken, refreshToken.Value)
	assert.Equal(t, false, refreshToken.Secure) // development
	assert.Equal(t, true, refreshToken.HttpOnly)
	assert.Equal(t, "/", refreshToken.Path)
}

func TestCookieHelper_ClearRefreshTokenCookie(t *testing.T) {
	tests := []struct {
		name  string
		env   string
		secure bool
	}{
		{
			name:  "Production environment",
			env:   "production",
			secure: true,
		},
		{
			name:  "Development environment",
			env:   "development",
			secure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{
					Env: tt.env,
				},
				JWT: config.JWTConfig{
					RefreshTokenExpireHours: 24,
				},
			}

			cookieHelper := helper.NewCookieHelper(cfg)
			app := fiber.New()

			// First set a cookie
			app.Post("/set", func(c *fiber.Ctx) error {
				cookieHelper.SetRefreshTokenCookie(c, "some-token")
				return c.SendString("set")
			})

			// Then clear it
			app.Post("/clear", func(c *fiber.Ctx) error {
				cookieHelper.ClearRefreshTokenCookie(c)
				return c.SendString("cleared")
			})

			// Set cookie
			req1, _ := http.NewRequest("POST", "/set", nil)
			resp1, err := app.Test(req1)
			assert.NoError(t, err)

			cookies1 := resp1.Cookies()
			var setCookie *http.Cookie
			for _, c := range cookies1 {
				if c.Name == "refresh_token" {
					setCookie = c
					break
				}
			}
			assert.NotNil(t, setCookie)
			assert.Equal(t, "some-token", setCookie.Value)

			// Clear cookie - the cleared cookie should have MaxAge < 0
			req2, _ := http.NewRequest("POST", "/clear", nil)
			resp2, err := app.Test(req2)
			assert.NoError(t, err)

			cookies2 := resp2.Cookies()
			var clearedCookie *http.Cookie
			for _, c := range cookies2 {
				if c.Name == "refresh_token" {
					clearedCookie = c
					break
				}
			}

			// Cleared cookie should have MaxAge = -1 and empty value
			assert.NotNil(t, clearedCookie)
			assert.Equal(t, "", clearedCookie.Value)
			assert.Equal(t, -1, clearedCookie.MaxAge)
		})
	}
}

func TestGetRefreshTokenFromCookie(t *testing.T) {
	app := fiber.New()

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
			app.Get("/test", func(c *fiber.Ctx) error {
				result := helper.GetRefreshTokenFromCookie(c)
				return c.SendString(result)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.token})
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)

			body := make([]byte, len(tt.expected))
			resp.Body.Read(body)
			assert.Equal(t, tt.expected, string(body))
		})
	}
}

func TestCookieHelper_SecurityProperties(t *testing.T) {
	t.Run("Production cookie has secure properties", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "production",
			},
			JWT: config.JWTConfig{
				RefreshTokenExpireHours: 72,
			},
		}

		cookieHelper := helper.NewCookieHelper(cfg)
		app := fiber.New()

		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "prod-token")
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", nil)
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
			JWT: config.JWTConfig{
				RefreshTokenExpireHours: 168, // 1 week
			},
		}

		cookieHelper := helper.NewCookieHelper(cfg)
		app := fiber.New()

		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "dev-token")
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", nil)
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
	t.Run("Special characters in token", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
			JWT: config.JWTConfig{
				RefreshTokenExpireHours: 24,
			},
		}

		cookieHelper := helper.NewCookieHelper(cfg)
		app := fiber.New()

		specialToken := "token.with.dots-and_underscores"
		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, specialToken)
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", nil)
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
			JWT: config.JWTConfig{
				RefreshTokenExpireHours: 24,
			},
		}

		cookieHelper := helper.NewCookieHelper(cfg)
		app := fiber.New()

		// Use URL-safe characters instead of raw unicode
		specialToken := "token-with-dashes_and_underscores-123"
		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, specialToken)
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", nil)
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
			JWT: config.JWTConfig{
				RefreshTokenExpireHours: 8760, // 1 year
			},
		}

		cookieHelper := helper.NewCookieHelper(cfg)
		app := fiber.New()

		app.Post("/test", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "long-expiry-token")
			return c.SendString("ok")
		})

		req, _ := http.NewRequest("POST", "/test", nil)
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
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 24,
		},
	}

	cookieHelper := helper.NewCookieHelper(cfg)
	app := fiber.New()

	t.Run("Complete cookie lifecycle", func(t *testing.T) {
		// Test setting a cookie
		app.Post("/set", func(c *fiber.Ctx) error {
			cookieHelper.SetRefreshTokenCookie(c, "test-token")
			return c.SendString("ok")
		})

		req1, _ := http.NewRequest("POST", "/set", nil)
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

		req2, _ := http.NewRequest("POST", "/clear", nil)
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
		assert.Equal(t, -1, cookie2.MaxAge)
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
		req1, _ := http.NewRequest("POST", "/token1", nil)
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

		req2, _ := http.NewRequest("POST", "/token2", nil)
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

func TestCookieHelper_MaxAgeCalculation(t *testing.T) {
	tests := []struct {
		name           string
		expireHours    int
		expectedMaxAge int
	}{
		{
			name:           "1 hour",
			expireHours:    1,
			expectedMaxAge: 3600,
		},
		{
			name:           "24 hours",
			expireHours:    24,
			expectedMaxAge: 86400,
		},
		{
			name:           "7 days",
			expireHours:    168,
			expectedMaxAge: 604800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{
					Env: "development",
				},
				JWT: config.JWTConfig{
					RefreshTokenExpireHours: tt.expireHours,
				},
			}

			cookieHelper := helper.NewCookieHelper(cfg)
			app := fiber.New()

			app.Post("/test", func(c *fiber.Ctx) error {
				cookieHelper.SetRefreshTokenCookie(c, "test-token")
				return c.SendString("ok")
			})

			req, _ := http.NewRequest("POST", "/test", nil)
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
			assert.Equal(t, "test-token", refreshToken.Value)
			assert.Equal(t, tt.expectedMaxAge, refreshToken.MaxAge)
		})
	}
}
