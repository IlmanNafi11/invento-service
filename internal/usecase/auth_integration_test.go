package usecase

import (
	"context"
	"invento-service/config"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type IntegrationMockAuthService struct {
	mock.Mock
}

func (m *IntegrationMockAuthService) Register(ctx context.Context, req domain.AuthServiceRegisterRequest) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *IntegrationMockAuthService) Login(ctx context.Context, email, password string) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *IntegrationMockAuthService) VerifyJWT(token string) (domain.AuthClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.AuthClaims), args.Error(1)
}

func (m *IntegrationMockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *IntegrationMockAuthService) RequestPasswordReset(ctx context.Context, email string, redirectTo string) error {
	args := m.Called(ctx, email, redirectTo)
	return args.Error(0)
}

func (m *IntegrationMockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *IntegrationMockAuthService) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

type integrationUserRepository struct {
	db *gorm.DB
}

func newIntegrationUserRepository(db *gorm.DB) *integrationUserRepository {
	return &integrationUserRepository{db: db}
}

func (r *integrationUserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ? AND is_active = ?", email, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *integrationUserRepository) GetByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ? AND is_active = ?", id, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *integrationUserRepository) GetProfileWithCounts(userID string) (*domain.User, int, int, error) {
	return nil, 0, 0, nil
}

func (r *integrationUserRepository) GetUserFiles(userID string, search string, page, limit int) ([]domain.UserFileItem, int, error) {
	return nil, 0, nil
}

func (r *integrationUserRepository) GetByIDs(userIDs []string) ([]*domain.User, error) {
	return nil, nil
}

func (r *integrationUserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *integrationUserRepository) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	return nil, 0, nil
}

func (r *integrationUserRepository) UpdateRole(userID string, roleID *int) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *integrationUserRepository) UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error {
	return nil
}

func (r *integrationUserRepository) Delete(userID string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

func (r *integrationUserRepository) GetByRoleID(roleID uint) ([]domain.UserListItem, error) {
	var users []domain.UserListItem
	err := r.db.Table("user_profiles").
		Select("id, email, name, role_id, is_active, created_at").
		Where("role_id = ? AND is_active = ?", roleID, true).
		Scan(&users).Error
	return users, err
}

func (r *integrationUserRepository) BulkUpdateRole(userIDs []string, roleID uint) error {
	return r.db.Model(&domain.User{}).Where("id IN ?", userIDs).Update("role_id", roleID).Error
}

type integrationRoleRepository struct {
	db *gorm.DB
}

func newIntegrationRoleRepository(db *gorm.DB) *integrationRoleRepository {
	return &integrationRoleRepository{db: db}
}

func (r *integrationRoleRepository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *integrationRoleRepository) GetByID(id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *integrationRoleRepository) GetByName(name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("nama_role = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *integrationRoleRepository) Update(role *domain.Role) error {
	return r.db.Save(role).Error
}

func (r *integrationRoleRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Role{}, id).Error
}

func (r *integrationRoleRepository) GetAll(search string, page, limit int) ([]domain.RoleListItem, int, error) {
	return nil, 0, nil
}

type IntegrationTestSuite struct {
	db          *gorm.DB
	userRepo    *integrationUserRepository
	roleRepo    *integrationRoleRepository
	mockAuth    *IntegrationMockAuthService
	authUsecase AuthUsecase
	cfg         *config.Config
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Create SQLite in-memory database for testing
	dsn := "file::memory:?cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Should connect to SQLite in-memory database")

	// Auto migrate test models
	err = db.AutoMigrate(&domain.Role{}, &domain.User{})
	require.NoError(t, err, "Should migrate tables successfully")

	// Create test config
	cfg := &config.Config{
		App: config.AppConfig{
			Env:            "development",
			CorsOriginDev:  "http://localhost:3000",
			CorsOriginProd: "https://example.com",
		},
		Supabase: config.SupabaseConfig{
			URL: "http://localhost:54321",
		},
	}

	// Create repositories
	userRepo := newIntegrationUserRepository(db)
	roleRepo := newIntegrationRoleRepository(db)

	// Create mock auth service
	mockAuth := new(IntegrationMockAuthService)

	// Create auth usecase with dependencies
	authUsecase := NewAuthUsecaseWithDeps(userRepo, roleRepo, mockAuth, cfg, zerolog.Nop())

	return &IntegrationTestSuite{
		db:          db,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		mockAuth:    mockAuth,
		authUsecase: authUsecase,
		cfg:         cfg,
	}
}

func (suite *IntegrationTestSuite) cleanup() {
	sqlDB, _ := suite.db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func (suite *IntegrationTestSuite) seedRole(name string) *domain.Role {
	role := &domain.Role{NamaRole: name}
	suite.db.Create(role)
	return role
}

func (suite *IntegrationTestSuite) seedUser(id, email, name string, roleID *int) *domain.User {
	user := &domain.User{
		ID:       id,
		Email:    email,
		Name:     name,
		RoleID:   roleID,
		IsActive: true,
	}
	suite.db.Create(user)
	return user
}

func TestAuthIntegration_RegisterFlow(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	mahasiswaRole := suite.seedRole("mahasiswa")

	t.Run("CompleteRegisterFlow_Success", func(t *testing.T) {
		// Setup
		supabaseUserID := "supabase-user-uuid-123"
		suite.mockAuth.On("Register", mock.Anything, mock.MatchedBy(func(r domain.AuthServiceRegisterRequest) bool {
			return r.Email == "test@student.polije.ac.id" && r.Password == "password123"
		})).Return(&domain.AuthServiceResponse{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User: &domain.AuthServiceUserInfo{
				ID:    supabaseUserID,
				Email: "test@student.polije.ac.id",
				Name:  "Test User",
			},
		}, nil).Once()

		// Execute: Call usecase
		req := domain.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		refreshToken, authResp, err := suite.authUsecase.Register(req)

		// Verify: No error and correct response
		require.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.Equal(t, "mock_refresh_token", refreshToken)
		assert.Equal(t, "mock_access_token", authResp.AccessToken)
		assert.Equal(t, "Test User", authResp.User.Name)
		assert.Equal(t, "test@student.polije.ac.id", authResp.User.Email)

		// Verify: User was created in database
		savedUser, err := suite.userRepo.GetByEmail("test@student.polije.ac.id")
		require.NoError(t, err)
		assert.Equal(t, supabaseUserID, savedUser.ID)
		assert.Equal(t, "Test User", savedUser.Name)
		assert.NotNil(t, savedUser.RoleID)
		assert.Equal(t, int(mahasiswaRole.ID), *savedUser.RoleID)
		assert.True(t, savedUser.IsActive)

		suite.mockAuth.AssertExpectations(t)
	})

	t.Run("RegisterFlow_EmailAlreadyExists", func(t *testing.T) {
		// Setup: Create existing user
		roleID := int(mahasiswaRole.ID)
		suite.seedUser("existing-user-id", "existing@student.polije.ac.id", "Existing User", &roleID)

		// Execute
		req := domain.RegisterRequest{
			Name:     "New User",
			Email:    "existing@student.polije.ac.id",
			Password: "password123",
		}

		refreshToken, authResp, err := suite.authUsecase.Register(req)

		// Verify: Should return conflict error
		require.Error(t, err)
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)

		var appErr *apperrors.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, fiber.StatusConflict, appErr.HTTPStatus)
	})

	t.Run("RegisterFlow_InvalidEmailDomain", func(t *testing.T) {
		// Execute
		req := domain.RegisterRequest{
			Name:     "Test User",
			Email:    "test@gmail.com", // Invalid domain
			Password: "password123",
		}

		refreshToken, authResp, err := suite.authUsecase.Register(req)

		// Verify: Should return validation error
		require.Error(t, err)
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "polije.ac.id")
	})
}

func TestAuthIntegration_LoginFlow(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Seed the role for students
	mahasiswaRole := suite.seedRole("mahasiswa")

	t.Run("CompleteLoginFlow_ExistingUser", func(t *testing.T) {
		// Setup: Create existing user in database
		roleID := int(mahasiswaRole.ID)
		suite.seedUser("existing-login-user", "login@student.polije.ac.id", "Login User", &roleID)

		// Setup: Mock Supabase auth service response
		suite.mockAuth.On("Login", mock.Anything, "login@student.polije.ac.id", "correctpassword").Return(&domain.AuthServiceResponse{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			User: &domain.AuthServiceUserInfo{
				ID:    "existing-login-user",
				Email: "login@student.polije.ac.id",
				Name:  "Login User",
			},
		}, nil).Once()

		// Execute
		req := domain.AuthRequest{
			Email:    "login@student.polije.ac.id",
			Password: "correctpassword",
		}

		refreshToken, authResp, err := suite.authUsecase.Login(req)

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.Equal(t, "mock_refresh_token", refreshToken)
		assert.Equal(t, "mock_access_token", authResp.AccessToken)
		assert.Equal(t, "Login User", authResp.User.Name)
		assert.Equal(t, "login@student.polije.ac.id", authResp.User.Email)

		suite.mockAuth.AssertExpectations(t)
	})

	t.Run("LoginFlow_InvalidCredentials", func(t *testing.T) {
		// Setup: Mock Supabase returning error for invalid credentials
		suite.mockAuth.On("Login", mock.Anything, "user@student.polije.ac.id", "wrongpassword").Return(nil, apperrors.NewUnauthorizedError("Invalid credentials")).Once()

		// Execute
		req := domain.AuthRequest{
			Email:    "user@student.polije.ac.id",
			Password: "wrongpassword",
		}

		refreshToken, authResp, err := suite.authUsecase.Login(req)

		// Verify
		require.Error(t, err)
		assert.Nil(t, authResp)
		assert.Empty(t, refreshToken)

		var appErr *apperrors.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, fiber.StatusUnauthorized, appErr.HTTPStatus)

		suite.mockAuth.AssertExpectations(t)
	})
}

func TestAuthIntegration_RequestPasswordReset(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Setup
	mahasiswaRole := suite.seedRole("mahasiswa")
	roleID := int(mahasiswaRole.ID)
	suite.seedUser("reset-user-id", "reset@student.polije.ac.id", "Reset User", &roleID)

	t.Run("RequestPasswordReset_Success", func(t *testing.T) {
		// Setup
		suite.mockAuth.On("RequestPasswordReset", mock.Anything, "reset@student.polije.ac.id", mock.Anything).Return(nil).Once()

		// Execute
		req := domain.ResetPasswordRequest{
			Email: "reset@student.polije.ac.id",
		}
		err := suite.authUsecase.RequestPasswordReset(req)

		// Verify
		require.NoError(t, err)
		suite.mockAuth.AssertExpectations(t)
	})

	t.Run("RequestPasswordReset_UnknownEmailStillDelegates", func(t *testing.T) {
		suite.mockAuth.On("RequestPasswordReset", mock.Anything, "nonexistent@student.polije.ac.id", mock.Anything).Return(nil).Once()

		req := domain.ResetPasswordRequest{Email: "nonexistent@student.polije.ac.id"}
		err := suite.authUsecase.RequestPasswordReset(req)

		require.NoError(t, err)
		suite.mockAuth.AssertExpectations(t)
	})
}

func TestAuthIntegration_RefreshAndLogout(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	t.Run("RefreshToken_Success", func(t *testing.T) {
		suite.mockAuth.On("RefreshToken", mock.Anything, "old_refresh").Return(&domain.AuthServiceResponse{
			AccessToken:  "new_access",
			RefreshToken: "new_refresh",
			TokenType:    "bearer",
			ExpiresIn:    3600,
		}, nil).Once()

		newRefresh, resp, err := suite.authUsecase.RefreshToken("old_refresh")
		require.NoError(t, err)
		assert.Equal(t, "new_refresh", newRefresh)
		assert.Equal(t, "new_access", resp.AccessToken)
		assert.NotZero(t, resp.ExpiresAt)
	})

	t.Run("Logout_Success", func(t *testing.T) {
		suite.mockAuth.On("Logout", mock.Anything, "access_token").Return(nil).Once()
		err := suite.authUsecase.Logout("access_token")
		require.NoError(t, err)
	})
}
