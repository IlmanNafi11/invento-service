package helper_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/helper"
	"invento-service/internal/supabase"
	testutil "invento-service/internal/testing"
	"invento-service/internal/usecase/repo"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthService implements domain.AuthService for testing
type mockAuthService struct {
	verifyJWTFunc func(token string) (domain.AuthClaims, error)
}

func (m *mockAuthService) Register(ctx context.Context, req domain.AuthServiceRegisterRequest) (*domain.AuthServiceResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.AuthServiceResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthService) VerifyJWT(token string) (domain.AuthClaims, error) {
	if m.verifyJWTFunc != nil {
		return m.verifyJWTFunc(token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthServiceResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAuthService) RequestPasswordReset(ctx context.Context, email string, redirectTo string) error {
	return errors.New("not implemented")
}

func (m *mockAuthService) Logout(ctx context.Context, accessToken string) error {
	return errors.New("not implemented")
}

func (m *mockAuthService) DeleteUser(ctx context.Context, uid string) error {
	return errors.New("not implemented")
}

func testSupabaseClaims(userID string) domain.AuthClaims {
	return &supabase.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
	}
}

func testCookieHelper() *helper.CookieHelper {
	return helper.NewCookieHelper(&config.Config{App: config.AppConfig{Env: "development"}})
}

// mockUserRepository implements repo.UserRepository for testing
type mockUserRepository struct {
	getByIDFunc func(id string) (*domain.User, error)
}

func (m *mockUserRepository) GetByEmail(email string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetByID(id string) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetProfileWithCounts(userID string) (*domain.User, int, int, error) {
	return nil, 0, 0, nil
}

func (m *mockUserRepository) GetUserFiles(userID string, search string, page, limit int) ([]domain.UserFileItem, int, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) GetByIDs(userIDs []string) ([]*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Create(user *domain.User) error {
	return errors.New("not implemented")
}

func (m *mockUserRepository) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *mockUserRepository) UpdateRole(userID string, roleID *int) error {
	return errors.New("not implemented")
}

func (m *mockUserRepository) UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error {
	return errors.New("not implemented")
}

func (m *mockUserRepository) Delete(userID string) error {
	return errors.New("not implemented")
}

func (m *mockUserRepository) GetByRoleID(roleID uint) ([]domain.UserListItem, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) BulkUpdateRole(userIDs []string, roleID uint) error {
	return errors.New("not implemented")
}

var _ repo.UserRepository = (*mockUserRepository)(nil)

func TestJWTAuthMiddleware_Creation(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}
	middleware := helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper())

	assert.NotNil(t, middleware)
}

func TestSupabaseAuthMiddleware_MissingAuthHeader_Returns401(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_InvalidTokenFormat_Returns401(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
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

func TestSupabaseAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	mockAuth := &mockAuthService{
		verifyJWTFunc: func(accessToken string) (domain.AuthClaims, error) {
			return nil, errors.New("invalid token")
		},
	}
	mockUser := &mockUserRepository{}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_ValidTokenWithUser_SetsContext(t *testing.T) {
	mockAuth := &mockAuthService{
		verifyJWTFunc: func(accessToken string) (domain.AuthClaims, error) {
			return testSupabaseClaims("user-123"), nil
		},
	}
	roleName := "admin"
	mockUser := &mockUserRepository{
		getByIDFunc: func(id string) (*domain.User, error) {
			return &domain.User{
				ID:       "user-123",
				Email:    "test@example.com",
				Name:     "Test User",
				IsActive: true,
				Role: &domain.Role{
					NamaRole: roleName,
				},
			}, nil
		},
	}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		userEmail := c.Locals("user_email").(string)
		userRole := c.Locals("user_role").(string)
		accessToken := c.Locals("access_token").(string)

		assert.Equal(t, "user-123", userID)
		assert.Equal(t, "test@example.com", userEmail)
		assert.Equal(t, "admin", userRole)
		assert.Equal(t, "valid-token", accessToken)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_UserNotFound_Returns401(t *testing.T) {
	mockAuth := &mockAuthService{
		verifyJWTFunc: func(accessToken string) (domain.AuthClaims, error) {
			return testSupabaseClaims("user-123"), nil
		},
	}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id string) (*domain.User, error) {
			return nil, errors.New("user not found")
		},
	}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_InactiveUser_Returns401(t *testing.T) {
	mockAuth := &mockAuthService{
		verifyJWTFunc: func(accessToken string) (domain.AuthClaims, error) {
			return testSupabaseClaims("user-123"), nil
		},
	}
	mockUser := &mockUserRepository{
		getByIDFunc: func(id string) (*domain.User, error) {
			return &domain.User{
				ID:       "user-123",
				Email:    "test@example.com",
				Name:     "Test User",
				Role:     nil,
				IsActive: false,
			}, nil
		},
	}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_ValidTokenFromCookie_Fallback(t *testing.T) {
	mockAuth := &mockAuthService{verifyJWTFunc: func(token string) (domain.AuthClaims, error) {
		if token != "cookie-token" {
			return nil, errors.New("unexpected token")
		}
		return testSupabaseClaims("user-123"), nil
	}}
	mockUser := &mockUserRepository{getByIDFunc: func(id string) (*domain.User, error) {
		return &domain.User{ID: id, Email: "test@example.com", IsActive: true}, nil
	}}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		assert.Equal(t, "cookie-token", c.Locals("access_token").(string))
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: helper.AccessTokenCookieName, Value: "cookie-token"})
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSupabaseAuthMiddleware_HeaderPrecedenceOverCookie(t *testing.T) {
	mockAuth := &mockAuthService{verifyJWTFunc: func(token string) (domain.AuthClaims, error) {
		if token != "header-token" {
			return nil, errors.New("header should take precedence")
		}
		return testSupabaseClaims("user-123"), nil
	}}
	mockUser := &mockUserRepository{getByIDFunc: func(id string) (*domain.User, error) {
		return &domain.User{ID: id, Email: "test@example.com", IsActive: true}, nil
	}}

	app := fiber.New()
	app.Use(helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		assert.Equal(t, "header-token", c.Locals("access_token").(string))
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	req.AddCookie(&http.Cookie{Name: helper.AccessTokenCookieName, Value: "cookie-token"})
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
	middleware := helper.TusProtocolMiddleware("1.0.0", 524288000)

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
			middleware := helper.TusProtocolMiddleware(version, 524288000)
			assert.NotNil(t, middleware)
		})
	}
}

func TestMiddleware_OptionsRequest(t *testing.T) {
	app := fiber.New()

	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))

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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	app.Use(helper.TusProtocolMiddleware("1.0.0", 524288000))
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
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}
	jwtMiddleware := helper.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper())
	rbacMiddleware := helper.RBACMiddleware(nil, "test", "read")
	tusMiddleware := helper.TusProtocolMiddleware("1.0.0", 524288000)

	assert.NotNil(t, jwtMiddleware)
	assert.NotNil(t, rbacMiddleware)
	assert.NotNil(t, tusMiddleware)
}

func TestTusProtocolMiddleware_HandlerType(t *testing.T) {
	middleware := helper.TusProtocolMiddleware("1.0.0", 524288000)

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
