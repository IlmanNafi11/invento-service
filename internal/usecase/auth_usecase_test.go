package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type AuthUsecaseMockAuthService struct {
	mock.Mock
}

func (m *AuthUsecaseMockAuthService) Register(ctx context.Context, req domain.AuthServiceRegisterRequest) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *AuthUsecaseMockAuthService) Login(ctx context.Context, email, password string) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *AuthUsecaseMockAuthService) VerifyJWT(token string) (domain.AuthClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.AuthClaims), args.Error(1)
}

func (m *AuthUsecaseMockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *AuthUsecaseMockAuthService) RequestPasswordReset(ctx context.Context, email string, redirectTo string) error {
	args := m.Called(ctx, email, redirectTo)
	return args.Error(0)
}

func (m *AuthUsecaseMockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *AuthUsecaseMockAuthService) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

type authTestUserRepo struct {
	mock.Mock
}

func (m *authTestUserRepo) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *authTestUserRepo) GetByID(id string) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *authTestUserRepo) GetProfileWithCounts(userID string) (*domain.User, int, int, error) {
	return nil, 0, 0, nil
}

func (m *authTestUserRepo) GetUserFiles(userID string, search string, page, limit int) ([]domain.UserFileItem, int, error) {
	return nil, 0, nil
}

func (m *authTestUserRepo) GetByIDs(userIDs []string) ([]*domain.User, error) {
	return nil, nil
}

func (m *authTestUserRepo) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *authTestUserRepo) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	args := m.Called(search, filterRole, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.UserListItem), args.Int(1), args.Error(2)
}

func (m *authTestUserRepo) UpdateRole(userID string, roleID *int) error {
	args := m.Called(userID, roleID)
	return args.Error(0)
}

func (m *authTestUserRepo) UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error {
	args := m.Called(userID, name, jenisKelamin, fotoProfil)
	return args.Error(0)
}

func (m *authTestUserRepo) Delete(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *authTestUserRepo) GetByRoleID(roleID uint) ([]domain.UserListItem, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserListItem), args.Error(1)
}

func (m *authTestUserRepo) BulkUpdateRole(userIDs []string, roleID uint) error {
	args := m.Called(userIDs, roleID)
	return args.Error(0)
}

type authTestRoleRepo struct {
	mock.Mock
}

func (m *authTestRoleRepo) Create(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *authTestRoleRepo) GetByID(id uint) (*domain.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *authTestRoleRepo) GetByName(name string) (*domain.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *authTestRoleRepo) Update(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *authTestRoleRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *authTestRoleRepo) GetAll(search string, page, limit int) ([]domain.RoleListItem, int, error) {
	args := m.Called(search, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.RoleListItem), args.Int(1), args.Error(2)
}

func newTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env:            "development",
			CorsOriginDev:  "http://localhost:3000",
			CorsOriginProd: "https://example.com",
		},
		Supabase: config.SupabaseConfig{
			URL: "http://localhost:54321",
		},
	}
}

func TestRegister_Success(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

	mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
	mockRole.On("GetByName", "mahasiswa").Return(role, nil)
	mockAuth.On("Register", mock.Anything, mock.MatchedBy(func(r domain.AuthServiceRegisterRequest) bool {
		return r.Email == req.Email && r.Password == req.Password && r.Name == req.Name
	})).Return(&domain.AuthServiceResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
		User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email, Name: req.Name},
	}, nil)
	mockUser.On("Create", mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == req.Email && u.Name == req.Name && u.ID == "user-uuid-123"
	})).Return(nil)

	refreshToken, authResp, err := uc.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.Equal(t, "refresh_token", refreshToken)
	assert.Equal(t, "access_token", authResp.AccessToken)
	assert.Equal(t, req.Email, authResp.User.Email)

	mockAuth.AssertCalled(t, "Register", mock.Anything, mock.Anything)
	mockUser.AssertCalled(t, "Create", mock.Anything)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRole.AssertExpectations(t)
}

func TestRegister_SupabaseFails(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

	mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
	mockRole.On("GetByName", "mahasiswa").Return(role, nil)
	mockAuth.On("Register", mock.Anything, mock.Anything).Return(nil, errors.New("supabase error"))

	refreshToken, authResp, err := uc.Register(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)

	mockAuth.AssertCalled(t, "Register", mock.Anything, mock.Anything)
	mockUser.AssertNotCalled(t, "Create", mock.Anything)
}

func TestRegister_LocalDBFails_RollbackSupabaseUser(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}
	supabaseUserID := "user-uuid-123"

	mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
	mockRole.On("GetByName", "mahasiswa").Return(role, nil)
	mockAuth.On("Register", mock.Anything, mock.Anything).Return(&domain.AuthServiceResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
		User:         &domain.AuthServiceUserInfo{ID: supabaseUserID, Email: req.Email, Name: req.Name},
	}, nil)
	mockUser.On("Create", mock.Anything).Return(errors.New("database error"))
	mockAuth.On("DeleteUser", mock.Anything, supabaseUserID).Return(nil)

	refreshToken, authResp, err := uc.Register(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)

	mockAuth.AssertCalled(t, "Register", mock.Anything, mock.Anything)
	mockUser.AssertCalled(t, "Create", mock.Anything)
	mockAuth.AssertCalled(t, "DeleteUser", mock.Anything, supabaseUserID)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	existingUser := &domain.User{ID: "existing-user", Email: req.Email}
	mockUser.On("GetByEmail", req.Email).Return(existingUser, nil)

	refreshToken, authResp, err := uc.Register(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "sudah terdaftar")

	mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestRegister_InvalidEmail(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@gmail.com",
		Password: "password123",
	}

	refreshToken, authResp, err := uc.Register(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "polije.ac.id")

	mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestLogin_Success_ExistingUser(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.AuthRequest{
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	roleID := 1
	existingUser := &domain.User{
		ID:       "user-uuid-123",
		Email:    req.Email,
		Name:     "Test User",
		RoleID:   &roleID,
		IsActive: true,
	}
	role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

	mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
		User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email, Name: "Test User"},
	}, nil)
	mockUser.On("GetByEmail", req.Email).Return(existingUser, nil)
	mockRole.On("GetByID", uint(1)).Return(role, nil)

	refreshToken, authResp, err := uc.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.Equal(t, "refresh_token", refreshToken)
	assert.Equal(t, "access_token", authResp.AccessToken)
	assert.Equal(t, req.Email, authResp.User.Email)

	mockAuth.AssertCalled(t, "Login", mock.Anything, req.Email, req.Password)
	mockUser.AssertCalled(t, "GetByEmail", req.Email)
}

func TestLogin_Success_NewUserSync(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.AuthRequest{
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
	mockUser.On("Create", mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == req.Email && u.ID == "new-user-uuid"
	})).Return(nil)
	mockRole.On("GetByID", uint(1)).Return(role, nil)

	refreshToken, authResp, err := uc.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.Equal(t, "refresh_token", refreshToken)
	assert.Equal(t, "access_token", authResp.AccessToken)

	mockAuth.AssertCalled(t, "Login", mock.Anything, req.Email, req.Password)
	mockUser.AssertCalled(t, "GetByEmail", req.Email)
	mockUser.AssertCalled(t, "Create", mock.Anything)
}

func TestLogin_SupabaseFails(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.AuthRequest{
		Email:    "test@student.polije.ac.id",
		Password: "wrongpassword",
	}

	mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(nil, errors.New("invalid credentials"))

	refreshToken, authResp, err := uc.Login(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "salah")

	mockAuth.AssertCalled(t, "Login", mock.Anything, req.Email, req.Password)
	mockUser.AssertNotCalled(t, "GetByEmail", mock.Anything)
}

func TestLogin_UserInactive(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.AuthRequest{
		Email:    "inactive@student.polije.ac.id",
		Password: "password123",
	}

	inactiveUser := &domain.User{
		ID:       "user-uuid-123",
		Email:    req.Email,
		Name:     "Inactive User",
		IsActive: false,
	}

	mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
		User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email},
	}, nil)
	mockUser.On("GetByEmail", req.Email).Return(inactiveUser, nil)

	refreshToken, authResp, err := uc.Login(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "belum diaktifkan")
}

func TestRequestPasswordReset_Success(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := domain.ResetPasswordRequest{
		Email: "test@student.polije.ac.id",
	}

	mockAuth.On("RequestPasswordReset", mock.Anything, req.Email, mock.Anything).Return(nil)

	err := uc.RequestPasswordReset(req)

	assert.NoError(t, err)
	mockAuth.AssertCalled(t, "RequestPasswordReset", mock.Anything, req.Email, mock.Anything)
}

func TestRefreshToken_Success(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	mockAuth.On("RefreshToken", mock.Anything, "refresh_token_old").Return(&domain.AuthServiceResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
	}, nil)

	newRefreshToken, resp, err := uc.RefreshToken("refresh_token_old")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "new_refresh_token", newRefreshToken)
	assert.Equal(t, "new_access_token", resp.AccessToken)
	assert.NotZero(t, resp.ExpiresAt)
}

func TestRefreshToken_ServiceFails(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	mockAuth.On("RefreshToken", mock.Anything, "refresh_token_old").Return(nil, errors.New("service error"))
	newRefreshToken, resp, err := uc.RefreshToken("refresh_token_old")

	assert.Error(t, err)
	assert.Empty(t, newRefreshToken)
	assert.Nil(t, resp)
}

func TestLogout_Success(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
	mockAuth.On("Logout", mock.Anything, "access_token").Return(nil)

	err := uc.Logout("access_token")
	assert.NoError(t, err)
	mockAuth.AssertCalled(t, "Logout", mock.Anything, "access_token")
}

func TestAuth_EdgeCases(t *testing.T) {
	t.Run("NewAuthUsecase_Constructor", func(t *testing.T) {
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()
		cfg.Supabase.JWTSecret = "test-secret"

		uc, err := NewAuthUsecase(mockUser, mockRole, nil, "service-key", cfg, zerolog.Nop())
		require.NoError(t, err)

		assert.NotNil(t, uc)
	})

	t.Run("Register_RoleNotFound", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(nil, gorm.ErrRecordNotFound)

		refreshToken, authResp, err := uc.Register(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Role mahasiswa")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("Register_RoleLookupGenericError", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(nil, errors.New("role query failed"))

		refreshToken, authResp, err := uc.Register(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("Register_SupabaseUserAlreadyRegistered", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}
		role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(role, nil)
		mockAuth.On("Register", mock.Anything, mock.Anything).Return(nil, errors.New("user already registered"))

		refreshToken, authResp, err := uc.Register(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email sudah terdaftar di sistem autentikasi")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
	})

	t.Run("Register_LocalCreateFailsWithNonAuthError", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}
		role := &domain.Role{ID: 1, NamaRole: "mahasiswa"}

		mockUser.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRole.On("GetByName", "mahasiswa").Return(role, nil)
		mockAuth.On("Register", mock.Anything, mock.Anything).Return(&domain.AuthServiceResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email, Name: req.Name},
		}, nil)
		mockUser.On("Create", mock.Anything).Return(errors.New("database write failed"))
		mockAuth.On("DeleteUser", mock.Anything, "user-uuid-123").Return(errors.New("delete failed"))

		refreshToken, authResp, err := uc.Register(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockAuth.AssertCalled(t, "DeleteUser", mock.Anything, "user-uuid-123")
	})

	t.Run("Login_GetByEmailReturnsGenericDBError", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.AuthRequest{
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

		refreshToken, authResp, err := uc.Login(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockRole.AssertNotCalled(t, "GetByName", mock.Anything)
	})

	t.Run("Login_AutoSyncInvalidEmailDomain", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.AuthRequest{
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

		refreshToken, authResp, err := uc.Login(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "polije.ac.id")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockRole.AssertNotCalled(t, "GetByName", mock.Anything)
	})

	t.Run("Login_AutoSyncNameFallbackToEmail", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.AuthRequest{
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
		mockUser.On("Create", mock.MatchedBy(func(u *domain.User) bool {
			return u.Name == req.Email
		})).Return(nil)

		refreshToken, authResp, err := uc.Login(req)

		assert.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.Equal(t, "refresh_token", refreshToken)
		assert.Equal(t, req.Email, authResp.User.Name)
	})

	t.Run("Login_AutoSyncRoleLookupFails", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.AuthRequest{
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

		refreshToken, authResp, err := uc.Login(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		mockUser.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("Login_AutoSyncCreateUserFails", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.AuthRequest{
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
		mockUser.On("Create", mock.Anything).Return(errors.New("insert failed"))

		refreshToken, authResp, err := uc.Login(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
	})

	t.Run("Logout_AuthServiceFails", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		mockAuth.On("Logout", mock.Anything, "access_token").Return(errors.New("logout failed"))

		err := uc.Logout("access_token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		mockAuth.AssertCalled(t, "Logout", mock.Anything, "access_token")
	})

	t.Run("RequestPasswordReset_AuthServiceFails", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.ResetPasswordRequest{Email: "test@student.polije.ac.id"}
		mockAuth.On("RequestPasswordReset", mock.Anything, req.Email, mock.Anything).Return(errors.New("request reset failed"))

		err := uc.RequestPasswordReset(req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")
		mockAuth.AssertCalled(t, "RequestPasswordReset", mock.Anything, req.Email, mock.Anything)
	})

	t.Run("RequestPasswordReset_ProductionRedirect", func(t *testing.T) {
		mockAuth := new(AuthUsecaseMockAuthService)
		mockUser := new(authTestUserRepo)
		mockRole := new(authTestRoleRepo)
		cfg := newTestConfig()
		cfg.App.Env = "production"

		uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
		req := domain.ResetPasswordRequest{Email: "test@student.polije.ac.id"}
		expectedRedirect := cfg.App.CorsOriginProd + "/reset-password"
		mockAuth.On("RequestPasswordReset", mock.Anything, req.Email, expectedRedirect).Return(nil)

		err := uc.RequestPasswordReset(req)

		assert.NoError(t, err)
		mockAuth.AssertCalled(t, "RequestPasswordReset", mock.Anything, req.Email, expectedRedirect)
	})
}
