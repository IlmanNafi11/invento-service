package helper_test

import (
	"bytes"
	"fiber-boiler-plate/internal/helper"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthMiddleware_Creation(t *testing.T) {
	// This tests the middleware creation - full testing would require JWT setup
	middleware := helper.JWTAuthMiddleware(nil)

	assert.NotNil(t, middleware)
}

func TestRBACMiddleware_Creation(t *testing.T) {
	// This tests the middleware creation - full testing would require Casbin setup
	middleware := helper.RBACMiddleware(nil, "resource", "action")

	assert.NotNil(t, middleware)
}

func TestTusProtocolMiddleware_Creation(t *testing.T) {
	// This tests the middleware creation
	middleware := helper.TusProtocolMiddleware("1.0.0")

	assert.NotNil(t, middleware)
}

func TestTusProtocolMiddleware_DifferentVersions(t *testing.T) {
	versions := []string{
		"1.0.0",
		"0.2.0",
		"1.1.0",
	}

	for _, version := range versions {
		t.Run("", func(t *testing.T) {
			middleware := helper.TusProtocolMiddleware(version)
			assert.NotNil(t, middleware)
		})
	}
}

func TestMiddleware_OptionsRequest(t *testing.T) {
	app := fiber.New()

	// Add TUS protocol middleware
	app.Use(helper.TusProtocolMiddleware("1.0.0"))

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/api/tus", bytes.NewBuffer(nil))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestMiddleware_HandlerCreation(t *testing.T) {
	// Test that all middleware creation functions return valid handlers
	jwtMiddleware := helper.JWTAuthMiddleware(nil)
	rbacMiddleware := helper.RBACMiddleware(nil, "test", "read")
	tusMiddleware := helper.TusProtocolMiddleware("1.0.0")

	assert.NotNil(t, jwtMiddleware)
	assert.NotNil(t, rbacMiddleware)
	assert.NotNil(t, tusMiddleware)
}

func TestTusProtocolMiddleware_HandlerType(t *testing.T) {
	middleware := helper.TusProtocolMiddleware("1.0.0")

	// Verify it's a fiber.Handler
	app := fiber.New()
	app.Use(middleware)

	// The app should be valid with the middleware attached
	assert.NotNil(t, app)
}

func TestRBACMiddleware_DifferentParameters(t *testing.T) {
	resources := []string{"projects", "users", "moduls"}
	actions := []string{"create", "read", "update", "delete"}

	for _, resource := range resources {
		for _, action := range actions {
			t.Run(resource+"_"+action, func(t *testing.T) {
				middleware := helper.RBACMiddleware(nil, resource, action)
				assert.NotNil(t, middleware)
			})
		}
	}
}
