package httputil_test

import (
	"invento-service/config"
	"invento-service/internal/httputil"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewCookieHelper(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}

	cookieHelper := httputil.NewCookieHelper(cfg)

	assert.NotNil(t, cookieHelper)
}

func TestCookieHelper_SetRefreshTokenCookie(t *testing.T) {
	t.Parallel()
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
			}

			cookieHelper := httputil.NewCookieHelper(cfg)
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
	t.Parallel()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}

	cookieHelper := httputil.NewCookieHelper(cfg)
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
	assert.Equal(t, httputil.RefreshTokenPath, refreshToken.Path)
}

func TestCookieHelper_ClearRefreshTokenCookie(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		env    string
		secure bool
	}{
		{
			name:   "Production environment",
			env:    "production",
			secure: true,
		},
		{
			name:   "Development environment",
			env:    "development",
			secure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{
					Env: tt.env,
				},
			}

			cookieHelper := httputil.NewCookieHelper(cfg)
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
			assert.Equal(t, 0, clearedCookie.MaxAge)
		})
	}
}

func TestCookieHelper_SetAccessTokenCookie(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{App: config.AppConfig{Env: "development"}}
	cookieHelper := httputil.NewCookieHelper(cfg)
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		cookieHelper.SetAccessTokenCookie(c, "access-token-value", 3600)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	var accessTokenCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == httputil.AccessTokenCookieName {
			accessTokenCookie = c
			break
		}
	}

	assert.NotNil(t, accessTokenCookie)
	assert.Equal(t, "access-token-value", accessTokenCookie.Value)
	assert.Equal(t, httputil.AccessTokenPath, accessTokenCookie.Path)
	assert.True(t, accessTokenCookie.HttpOnly)
}

func TestCookieHelper_MaxAgeCalculation(t *testing.T) {
	t.Parallel()
	expectedMaxAge := 0

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}

	cookieHelper := httputil.NewCookieHelper(cfg)
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
	assert.Equal(t, expectedMaxAge, refreshToken.MaxAge)
}
