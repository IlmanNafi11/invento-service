package middleware_test

import (
	"bytes"
	"invento-service/internal/middleware"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTusProtocolMiddleware_Creation(t *testing.T) {
	mw := middleware.TusProtocolMiddleware("1.0.0", 524288000)

	assert.NotNil(t, mw)
}

func TestTusProtocolMiddleware_DifferentVersions(t *testing.T) {
	versions := []string{
		"1.0.0",
		"0.2.0",
		"1.1.0",
	}

	for _, version := range versions {
		t.Run("", func(t *testing.T) {
			mw := middleware.TusProtocolMiddleware(version, 524288000)
			assert.NotNil(t, mw)
		})
	}
}

func TestMiddleware_OptionsRequest(t *testing.T) {
	app := fiber.New()

	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))

	req := httptest.NewRequest("OPTIONS", "/api/tus", bytes.NewBuffer(nil))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Version"))
	assert.Equal(t, "creation,termination", resp.Header.Get("Tus-Extension"))
	assert.Equal(t, "524288000", resp.Header.Get("Tus-Max-Size"))
}

func TestTusProtocolMiddleware_MissingTusResumableOnPatch_Returns412(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusPreconditionFailed, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
}

func TestTusProtocolMiddleware_WrongTusVersion_Returns412(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	req.Header.Set("Tus-Resumable", "0.9.0")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusPreconditionFailed, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
}

func TestTusProtocolMiddleware_GetRequestWithoutTusHeader_Passes(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTusProtocolMiddleware_PostRequestWithoutTusHeader_Passes(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTusProtocolMiddleware_ValidTusVersion_Passes(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTusProtocolMiddleware_HeadRequestWithTusHeader_Passes(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Head("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTusProtocolMiddleware_DeleteRequestWithTusHeader_Passes(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TusProtocolMiddleware("1.0.0", 524288000))
	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTusProtocolMiddleware_HandlerType(t *testing.T) {
	mw := middleware.TusProtocolMiddleware("1.0.0", 524288000)

	app := fiber.New()
	app.Use(mw)

	assert.NotNil(t, app)
}
