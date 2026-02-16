package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *AuthUsecaseMockAuthService) RequestPasswordReset(ctx context.Context, email, redirectTo string) error {
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

func (m *authTestUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *authTestUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *authTestUserRepo) GetProfileWithCounts(ctx context.Context, userID string) (user *domain.User, projectCount, modulCount int, err error) {
	return nil, 0, 0, nil
}

func (m *authTestUserRepo) GetUserFiles(ctx context.Context, userID, search string, page, limit int) ([]dto.UserFileItem, int, error) {
	return nil, 0, nil
}

func (m *authTestUserRepo) GetByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	return nil, nil
}

func (m *authTestUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *authTestUserRepo) GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]dto.UserListItem, int, error) {
	args := m.Called(search, filterRole, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]dto.UserListItem), args.Int(1), args.Error(2)
}

func (m *authTestUserRepo) UpdateRole(ctx context.Context, userID string, roleID *int) error {
	args := m.Called(userID, roleID)
	return args.Error(0)
}

func (m *authTestUserRepo) UpdateProfile(ctx context.Context, userID, name string, jenisKelamin, fotoProfil *string) error {
	args := m.Called(userID, name, jenisKelamin, fotoProfil)
	return args.Error(0)
}

func (m *authTestUserRepo) Delete(ctx context.Context, userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *authTestUserRepo) GetByRoleID(ctx context.Context, roleID uint) ([]dto.UserListItem, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.UserListItem), args.Error(1)
}

func (m *authTestUserRepo) BulkUpdateRole(ctx context.Context, userIDs []string, roleID uint) error {
	args := m.Called(userIDs, roleID)
	return args.Error(0)
}

type authTestRoleRepo struct {
	mock.Mock
}

func (m *authTestRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *authTestRoleRepo) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *authTestRoleRepo) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *authTestRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *authTestRoleRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *authTestRoleRepo) GetAll(ctx context.Context, search string, page, limit int) ([]dto.RoleListItem, int, error) {
	args := m.Called(search, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]dto.RoleListItem), args.Int(1), args.Error(2)
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
	t.Parallel()
	mockAuth := new(MockAuthService)
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

	refreshToken, authResp, err := uc.Register(context.Background(), req)

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
	t.Parallel()
	mockAuth := new(MockAuthService)
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
	mockAuth.On("Register", mock.Anything, mock.Anything).Return(nil, errors.New("supabase error"))

	refreshToken, authResp, err := uc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)

	mockAuth.AssertCalled(t, "Register", mock.Anything, mock.Anything)
	mockUser.AssertNotCalled(t, "Create", mock.Anything)
}

func TestRegister_LocalDBFails_RollbackSupabaseUser(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
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

	refreshToken, authResp, err := uc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)

	mockAuth.AssertCalled(t, "Register", mock.Anything, mock.Anything)
	mockUser.AssertCalled(t, "Create", mock.Anything)
	mockAuth.AssertCalled(t, "DeleteUser", mock.Anything, supabaseUserID)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	existingUser := &domain.User{ID: "existing-user", Email: req.Email}
	mockUser.On("GetByEmail", req.Email).Return(existingUser, nil)

	refreshToken, authResp, err := uc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "sudah terdaftar")

	mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestRegister_InvalidEmail(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@gmail.com",
		Password: "password123",
	}

	refreshToken, authResp, err := uc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "polije.ac.id")

	mockAuth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestLogin_Success_ExistingUser(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := dto.AuthRequest{
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

	refreshToken, authResp, err := uc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.Equal(t, "refresh_token", refreshToken)
	assert.Equal(t, "access_token", authResp.AccessToken)
	assert.Equal(t, req.Email, authResp.User.Email)

	mockAuth.AssertCalled(t, "Login", mock.Anything, req.Email, req.Password)
	mockUser.AssertCalled(t, "GetByEmail", req.Email)
}

func TestLogin_Success_NewUserSync(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
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
	mockUser.On("Create", mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == req.Email && u.ID == "new-user-uuid"
	})).Return(nil)
	mockRole.On("GetByID", uint(1)).Return(role, nil)

	refreshToken, authResp, err := uc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.Equal(t, "refresh_token", refreshToken)
	assert.Equal(t, "access_token", authResp.AccessToken)

	mockAuth.AssertCalled(t, "Login", mock.Anything, req.Email, req.Password)
	mockUser.AssertCalled(t, "GetByEmail", req.Email)
	mockUser.AssertCalled(t, "Create", mock.Anything)
}

func TestLogin_SupabaseFails(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := dto.AuthRequest{
		Email:    "test@student.polije.ac.id",
		Password: "wrongpassword",
	}

	mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(nil, errors.New("invalid credentials"))

	refreshToken, authResp, err := uc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "salah")

	mockAuth.AssertCalled(t, "Login", mock.Anything, req.Email, req.Password)
	mockUser.AssertNotCalled(t, "GetByEmail", mock.Anything)
}
