package middleware_test

import (
	"encoding/json"
	"errors"
	"invento-service/internal/middleware"
	testutil "invento-service/internal/testing"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockErrorEnforcer struct{}

func (m *mockErrorEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	return false, errors.New("casbin error")
}

func TestRBACMiddleware_Creation(t *testing.T) {
	mw := middleware.RBACMiddleware(nil, "resource", "action")

	assert.NotNil(t, mw)
}

func TestRBACMiddleware_ValidRoleWithPermission_Returns200(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcerWithPolicies([][]string{
		{"admin", "projects", "read"},
	})
	require.NoError(t, err)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", "admin")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRBACMiddleware_ValidRoleWithoutPermission_Returns403(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcerWithPolicies([][]string{
		{"admin", "projects", "read"},
	})
	require.NoError(t, err)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", "admin")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "delete"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "error", body["status"])
}

func TestRBACMiddleware_InvalidRole_Returns403(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcerWithPolicies([][]string{
		{"admin", "projects", "read"},
	})
	require.NoError(t, err)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", "guest")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_MissingRoleInContext_Returns403(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcer()
	require.NoError(t, err)

	app := fiber.New()
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_EmptyRoleInContext_Returns403(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcer()
	require.NoError(t, err)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", "")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_InvalidRoleType_Returns403(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcer()
	require.NoError(t, err)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", 12345)
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRBACMiddleware_CasbinError_Returns500(t *testing.T) {
	enforcer := &mockErrorEnforcer{}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", "admin")
		return c.Next()
	})
	app.Use(middleware.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestRBACMiddleware_MultiplePermissions(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcerWithPolicies([][]string{
		{"admin", "projects", "create"},
		{"admin", "projects", "read"},
		{"admin", "projects", "update"},
		{"admin", "projects", "delete"},
		{"user", "projects", "read"},
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		role           string
		resource       string
		action         string
		expectedStatus int
	}{
		{"admin can create", "admin", "projects", "create", fiber.StatusOK},
		{"admin can read", "admin", "projects", "read", fiber.StatusOK},
		{"admin can update", "admin", "projects", "update", fiber.StatusOK},
		{"admin can delete", "admin", "projects", "delete", fiber.StatusOK},
		{"user can read", "user", "projects", "read", fiber.StatusOK},
		{"user cannot create", "user", "projects", "create", fiber.StatusForbidden},
		{"user cannot delete", "user", "projects", "delete", fiber.StatusForbidden},
		{"guest cannot read", "guest", "projects", "read", fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				return c.Next()
			})
			app.Use(middleware.RBACMiddleware(enforcer, tt.resource, tt.action))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRBACMiddleware_DifferentResources(t *testing.T) {
	enforcer, err := testutil.NewTestCasbinEnforcerWithPolicies([][]string{
		{"admin", "projects", "read"},
		{"admin", "users", "read"},
		{"user", "projects", "read"},
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		role           string
		resource       string
		action         string
		expectedStatus int
	}{
		{"admin can read projects", "admin", "projects", "read", fiber.StatusOK},
		{"admin can read users", "admin", "users", "read", fiber.StatusOK},
		{"user can read projects", "user", "projects", "read", fiber.StatusOK},
		{"user cannot read users", "user", "users", "read", fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				return c.Next()
			})
			app.Use(middleware.RBACMiddleware(enforcer, tt.resource, tt.action))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRBACMiddleware_DifferentParameters(t *testing.T) {
	resources := []string{"projects", "users", "moduls"}
	actions := []string{"create", "read", "update", "delete"}

	for _, resource := range resources {
		for _, action := range actions {
			t.Run(resource+"_"+action, func(t *testing.T) {
				mw := middleware.RBACMiddleware(nil, resource, action)
				assert.NotNil(t, mw)
			})
		}
	}
}
