package middleware_test

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/supabase"
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

func testCookieHelper() *httputil.CookieHelper {
	return httputil.NewCookieHelper(&config.Config{App: config.AppConfig{Env: "development"}})
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

func (m *mockUserRepository) GetUserFiles(userID string, search string, page, limit int) ([]dto.UserFileItem, int, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) GetByIDs(userIDs []string) ([]*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Create(user *domain.User) error {
	return errors.New("not implemented")
}

func (m *mockUserRepository) GetAll(search, filterRole string, page, limit int) ([]dto.UserListItem, int, error) {
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

func (m *mockUserRepository) GetByRoleID(roleID uint) ([]dto.UserListItem, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) BulkUpdateRole(userIDs []string, roleID uint) error {
	return errors.New("not implemented")
}

var _ repo.UserRepository = (*mockUserRepository)(nil)

func TestJWTAuthMiddleware_Creation(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}
	mw := middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper())

	assert.NotNil(t, mw)
}

func TestSupabaseAuthMiddleware_MissingAuthHeader_Returns401(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}

	app := fiber.New()
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// test fixtures - not real credentials
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		assert.Equal(t, "cookie-token", c.Locals("access_token").(string))
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: httputil.AccessTokenCookieName, Value: "cookie-token"})
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
	app.Use(middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper()))
	app.Get("/test", func(c *fiber.Ctx) error {
		assert.Equal(t, "header-token", c.Locals("access_token").(string))
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	req.AddCookie(&http.Cookie{Name: httputil.AccessTokenCookieName, Value: "cookie-token"})
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_HandlerCreation(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockUser := &mockUserRepository{}
	jwtMiddleware := middleware.SupabaseAuthMiddleware(mockAuth, mockUser, testCookieHelper())
	rbacMiddleware := middleware.RBACMiddleware(nil, "test", "read")
	tusMiddleware := middleware.TusProtocolMiddleware("1.0.0", 524288000)

	assert.NotNil(t, jwtMiddleware)
	assert.NotNil(t, rbacMiddleware)
	assert.NotNil(t, tusMiddleware)
}
