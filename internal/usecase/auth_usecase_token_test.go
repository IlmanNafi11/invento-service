package usecase

import (
	"context"
	"errors"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuth_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("NewAuthUsecase_Constructor", func(t *testing.T) {
		t.Skip("requires network access to Supabase JWKS endpoint")
		t.Parallel()
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()
		cfg.Supabase.JWTSecret = "test-secret"

		uc, err := NewAuthUsecase(mockUser, mockRole, nil, "service-key", cfg, zerolog.Nop())
		require.NoError(t, err)

		assert.NotNil(t, uc)
	})

	t.Run("Register_RoleNotFound", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(nil, gorm.ErrRecordNotFound)

		result, err := uc.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Role mahasiswa")
		assert.Nil(t, result)
		mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("Register_RoleLookupGenericError", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(nil, errors.New("role query failed"))

		result, err := uc.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, result)
		mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("Register_SupabaseUserAlreadyRegistered", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}
		role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(role, nil)
		mockAuth.On("Register", mock.Anything, mock.Anything).Return(nil, errors.New("user already registered"))

		result, err := uc.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email sudah terdaftar di sistem autentikasi")
		assert.Nil(t, result)
	})

	t.Run("Register_LocalCreateFailsWithNonAuthError", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}
		role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(role, nil)
		mockAuth.On("Register", mock.Anything, mock.Anything).Return(&domain.AuthServiceResponse{
			AccessToken:  "",
			RefreshToken: "",
			TokenType:    "bearer",
			ExpiresIn:    0,
			User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email, Name: req.Name},
		}, nil)
		mockUser.On("SaveOrUpdate", mock.Anything).Return(errors.New("database write failed"))
		mockAuth.On("DeleteUser", mock.Anything, "user-uuid-123").Return(errors.New("delete failed"))

		result, err := uc.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, result)
		mockAuth.AssertCalled(t, "DeleteUser", mock.Anything, "user-uuid-123")
	})

	t.Run("Login_GetByEmailReturnsGenericDBError", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.AuthRequest{
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email, Name: "Test User"},
		}, nil)
		mockUser.On("GetByEmail", req.Email).Return(nil, errors.New("database connection failed"))

		refreshToken, authResp, err := uc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockRole.AssertNotCalled(t, "GetByName", mock.Anything)
	})

	t.Run("Login_AutoSyncInvalidEmailDomain", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.AuthRequest{
			Email:    "invalid@gmail.com",
			Password: "password123",
		}

		mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User:         &domain.AuthServiceUserInfo{ID: "new-user-uuid", Email: req.Email, Name: "New User"},
		}, nil)
		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)

		refreshToken, authResp, err := uc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "polije.ac.id")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockRole.AssertNotCalled(t, "GetByName", mock.Anything)
	})

	t.Run("Login_AutoSyncNameFallbackToEmail", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.AuthRequest{
			Email:    "fallback@student.polije.ac.id",
			Password: "password123",
		}
		role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

		mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User:         &domain.AuthServiceUserInfo{ID: "new-user-uuid", Email: req.Email, Name: ""},
		}, nil)
		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(role, nil)
		mockUser.On("SaveOrUpdate", mock.MatchedBy(func(u *domain.User) bool {
			return u.Name == req.Email
		})).Return(nil)

		refreshToken, authResp, err := uc.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.Equal(t, "refresh_token", refreshToken)
		assert.Equal(t, req.Email, authResp.User.Name)
	})

	t.Run("Login_AutoSyncRoleLookupFails", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.AuthRequest{
			Email:    "newuser@student.polije.ac.id",
			Password: "password123",
		}

		mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User:         &domain.AuthServiceUserInfo{ID: "new-user-uuid", Email: req.Email, Name: "New User"},
		}, nil)
		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(nil, errors.New("role lookup failed"))

		refreshToken, authResp, err := uc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockUser.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("Login_AutoSyncCreateUserFails", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.AuthRequest{
			Email:    "newuser@student.polije.ac.id",
			Password: "password123",
		}
		role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

		mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User:         &domain.AuthServiceUserInfo{ID: "new-user-uuid", Email: req.Email, Name: "New User"},
		}, nil)
		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(role, nil)
		mockUser.On("SaveOrUpdate", mock.Anything).Return(errors.New("insert failed"))

		refreshToken, authResp, err := uc.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
	})

	t.Run("Logout_AuthServiceFails", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		mockAuth.On("Logout", mock.Anything, "access_token").Return(errors.New("logout failed"))

		err := uc.Logout(context.Background(), "access_token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		mockAuth.AssertCalled(t, "Logout", mock.Anything, "access_token")
	})

	t.Run("RequestPasswordReset_AuthServiceFails", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.ResetPasswordRequest{Email: "test@student.polije.ac.id"}
		mockAuth.On("RequestPasswordReset", mock.Anything, req.Email, mock.Anything).Return(errors.New("request reset failed"))

		err := uc.RequestPasswordReset(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		mockAuth.AssertCalled(t, "RequestPasswordReset", mock.Anything, req.Email, mock.Anything)
	})

	t.Run("RequestPasswordReset_ProductionRedirect", func(t *testing.T) {
		t.Parallel()
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()
		cfg.App.Env = "production"

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := dto.ResetPasswordRequest{Email: "test@student.polije.ac.id"}
		expectedRedirect := cfg.App.CorsOriginProd + "/reset-password"
		mockAuth.On("RequestPasswordReset", mock.Anything, req.Email, expectedRedirect).Return(nil)

		err := uc.RequestPasswordReset(context.Background(), req)

		assert.NoError(t, err)
		mockAuth.AssertCalled(t, "RequestPasswordReset", mock.Anything, req.Email, expectedRedirect)
	})
}
