package helper_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fiber-boiler-plate/internal/helper"
	testutil "fiber-boiler-plate/internal/testing"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthMiddleware_Creation(t *testing.T) {
	middleware := helper.SupabaseAuthMiddleware("")

	assert.NotNil(t, middleware)
}

func TestSupabaseAuthMiddleware_MissingAuthHeader_Returns401(t *testing.T) {
	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware("https://example.supabase.co"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_InvalidTokenFormat_Returns401(t *testing.T) {
	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware("https://example.supabase.co"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	tests := []struct {
		name       string
		authHeader string
	}{
		{"missing bearer prefix", "invalid-token"},
		{"wrong prefix", "Basic token123"},
		{"only bearer", "Bearer"},
		{"too many parts", "Bearer token extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		})
	}
}

func TestSupabaseAuthMiddleware_ValidBearerToken_SetsLocalAndContinues(t *testing.T) {
	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware("https://example.supabase.co"))
	app.Get("/test", func(c *fiber.Ctx) error {
		token := c.Locals("access_token")
		assert.Equal(t, "valid-test-token-123", token)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-test-token-123")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRBACMiddleware_Creation(t *testing.T) {
	middleware := helper.RBACMiddleware(nil, "resource", "action")

	assert.NotNil(t, middleware)
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
	app.Use(helper.RBACMiddleware(enforcer, "projects", "read"))
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
	app.Use(helper.RBACMiddleware(enforcer, "projects", "delete"))
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
	assert.Equal(t, false, body["success"])
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
	app.Use(helper.RBACMiddleware(enforcer, "projects", "read"))
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
	app.Use(helper.RBACMiddleware(enforcer, "projects", "read"))
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
	app.Use(helper.RBACMiddleware(enforcer, "projects", "read"))
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
	app.Use(helper.RBACMiddleware(enforcer, "projects", "read"))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

type mockErrorEnforcer struct{}

func (m *mockErrorEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	return false, errors.New("casbin error")
}

func TestRBACMiddleware_CasbinError_Returns500(t *testing.T) {
	enforcer := &mockErrorEnforcer{}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", "admin")
		return c.Next()
	})
	app.Use(helper.RBACMiddleware(enforcer, "projects", "read"))
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
			app.Use(helper.RBACMiddleware(enforcer, tt.resource, tt.action))
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
			app.Use(helper.RBACMiddleware(enforcer, tt.resource, tt.action))
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

func TestTusProtocolMiddleware_Creation(t *testing.T) {
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

	app.Use(helper.TusProtocolMiddleware("1.0.0"))

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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0"))
	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_HandlerCreation(t *testing.T) {
	jwtMiddleware := helper.SupabaseAuthMiddleware("")
	rbacMiddleware := helper.RBACMiddleware(nil, "test", "read")
	tusMiddleware := helper.TusProtocolMiddleware("1.0.0")

	assert.NotNil(t, jwtMiddleware)
	assert.NotNil(t, rbacMiddleware)
	assert.NotNil(t, tusMiddleware)
}

func TestTusProtocolMiddleware_HandlerType(t *testing.T) {
	middleware := helper.TusProtocolMiddleware("1.0.0")

	app := fiber.New()
	app.Use(middleware)

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
