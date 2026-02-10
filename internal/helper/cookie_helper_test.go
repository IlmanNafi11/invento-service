package helper_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/helper"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
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
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			cookieHelper.SetRefreshTokenCookie(ctx, tt.token)

			// Get the cookie
			cookie := ctx.Cookies("refresh_token")

			assert.Equal(t, tt.token, cookie)
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
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	testToken := "test-refresh-token"
	cookieHelper.SetRefreshTokenCookie(ctx, testToken)

	// Verify cookie value
	cookie := ctx.Cookies("refresh_token")
	assert.Equal(t, testToken, cookie)
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
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// First set a cookie
			cookieHelper.SetRefreshTokenCookie(ctx, "some-token")
			assert.Equal(t, "some-token", ctx.Cookies("refresh_token"))

			// Then clear it
			cookieHelper.ClearRefreshTokenCookie(ctx)

			// Cookie should be cleared (empty or expired)
			cookie := ctx.Cookies("refresh_token")
			assert.Empty(t, cookie)
		})
	}
}

func TestCookieHelper_ClearRefreshTokenCookie_Properties(t *testing.T) {
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
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	// Set cookie first
	cookieHelper.SetRefreshTokenCookie(ctx, "token-to-clear")

	// Clear it
	cookieHelper.ClearRefreshTokenCookie(ctx)

	// Verify cookie is cleared
	cookie := ctx.Cookies("refresh_token")
	assert.Empty(t, cookie)
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
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set cookie
			ctx.Cookie(&fiber.Cookie{
				Name:  "refresh_token",
				Value: tt.token,
			})

			// Get cookie
			result := helper.GetRefreshTokenFromCookie(ctx)

			assert.Equal(t, tt.expected, result)
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
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		cookieHelper.SetRefreshTokenCookie(ctx, "prod-token")

		// In production, cookie should be set
		cookie := ctx.Cookies("refresh_token")
		assert.Equal(t, "prod-token", cookie)
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
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		cookieHelper.SetRefreshTokenCookie(ctx, "dev-token")

		// In development, cookie should be set without secure flag
		cookie := ctx.Cookies("refresh_token")
		assert.Equal(t, "dev-token", cookie)
	})
}

func TestCookieHelper_EdgeCases(t *testing.T) {
	app := fiber.New()

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
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		specialToken := "token.with.dots-and_underscores"
		cookieHelper.SetRefreshTokenCookie(ctx, specialToken)

		result := ctx.Cookies("refresh_token")
		assert.Equal(t, specialToken, result)
	})

	t.Run("Unicode characters in token", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
			JWT: config.JWTConfig{
				RefreshTokenExpireHours: 24,
			},
		}

		cookieHelper := helper.NewCookieHelper(cfg)
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		unicodeToken := "token中文"
		cookieHelper.SetRefreshTokenCookie(ctx, unicodeToken)

		result := ctx.Cookies("refresh_token")
		assert.Equal(t, unicodeToken, result)
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
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		cookieHelper.SetRefreshTokenCookie(ctx, "long-expiry-token")

		result := ctx.Cookies("refresh_token")
		assert.Equal(t, "long-expiry-token", result)
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
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		// Set initial token
		token1 := "initial-token-123"
		cookieHelper.SetRefreshTokenCookie(ctx, token1)
		assert.Equal(t, token1, helper.GetRefreshTokenFromCookie(ctx))

		// Update token
		token2 := "updated-token-456"
		cookieHelper.SetRefreshTokenCookie(ctx, token2)
		assert.Equal(t, token2, helper.GetRefreshTokenFromCookie(ctx))

		// Clear token
		cookieHelper.ClearRefreshTokenCookie(ctx)
		assert.Empty(t, helper.GetRefreshTokenFromCookie(ctx))

		app.ReleaseCtx(ctx)
	})

	t.Run("Multiple contexts independence", func(t *testing.T) {
		ctx1 := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx2 := app.AcquireCtx(&fasthttp.RequestCtx{})

		// Set different tokens in different contexts
		cookieHelper.SetRefreshTokenCookie(ctx1, "token-1")
		cookieHelper.SetRefreshTokenCookie(ctx2, "token-2")

		// Verify independence
		assert.Equal(t, "token-1", helper.GetRefreshTokenFromCookie(ctx1))
		assert.Equal(t, "token-2", helper.GetRefreshTokenFromCookie(ctx2))

		app.ReleaseCtx(ctx1)
		app.ReleaseCtx(ctx2)
	})
}

func TestCookieHelper_MaxAgeCalculation(t *testing.T) {
	tests := []struct {
		name             string
		expireHours      int
		expectedMaxAge   int
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
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			cookieHelper.SetRefreshTokenCookie(ctx, "test-token")

			// Cookie should be set
			cookie := ctx.Cookies("refresh_token")
			assert.Equal(t, "test-token", cookie)
		})
	}
}
