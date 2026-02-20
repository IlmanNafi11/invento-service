package usecase

import (
	"context"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"testing"

	apperrors "invento-service/internal/errors"

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

func (m *IntegrationMockAuthService) RequestPasswordReset(ctx context.Context, email, redirectTo string) error {
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

func (m *IntegrationMockAuthService) ResendConfirmation(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *IntegrationMockAuthService) AdminCreateUser(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

type integrationUserRepository struct {
	db *gorm.DB
}

func newIntegrationUserRepository(db *gorm.DB) *integrationUserRepository {
	return &integrationUserRepository{db: db}
}

func (r *integrationUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ? AND is_active = ?", email, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *integrationUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *integrationUserRepository) GetProfileWithCounts(ctx context.Context, userID string) (user *domain.User, projectCount, modulCount int, err error) {
	return nil, 0, 0, nil
}

func (r *integrationUserRepository) GetUserFiles(ctx context.Context, userID, search string, page, limit int) ([]dto.UserFileItem, int, error) {
	return nil, 0, nil
}

func (r *integrationUserRepository) GetByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	return nil, nil
}

func (r *integrationUserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *integrationUserRepository) SaveOrUpdate(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *integrationUserRepository) GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]dto.UserListItem, int, error) {
	return nil, 0, nil
}

func (r *integrationUserRepository) UpdateRole(ctx context.Context, userID string, roleID *int) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *integrationUserRepository) UpdateProfile(ctx context.Context, userID, name string, jenisKelamin, fotoProfil *string) error {
	return nil
}

func (r *integrationUserRepository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

func (r *integrationUserRepository) GetByRoleID(ctx context.Context, roleID uint) ([]dto.UserListItem, error) {
	var users []dto.UserListItem
	err := r.db.WithContext(ctx).Table("user_profiles").
		Select("id, email, name, role_id, is_active, created_at").
		Where("role_id = ? AND is_active = ?", roleID, true).
		Scan(&users).Error
	return users, err
}

func (r *integrationUserRepository) BulkUpdateRole(ctx context.Context, userIDs []string, roleID uint) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id IN ?", userIDs).Update("role_id", roleID).Error
}

func (r *integrationUserRepository) FindByEmails(ctx context.Context, emails []string) ([]domain.User, error) {
	var users []domain.User
	if len(emails) == 0 {
		return users, nil
	}
	err := r.db.WithContext(ctx).Where("email IN ?", emails).Find(&users).Error
	return users, err
}

type integrationRoleRepository struct {
	db *gorm.DB
}

func newIntegrationRoleRepository(db *gorm.DB) *integrationRoleRepository {
	return &integrationRoleRepository{db: db}
}

func (r *integrationRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *integrationRoleRepository) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *integrationRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where("nama_role = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *integrationRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *integrationRoleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, id).Error
}

func (r *integrationRoleRepository) GetAll(ctx context.Context, search string, page, limit int) ([]dto.RoleListItem, int, error) {
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

func (suite *IntegrationTestSuite) seedUser(id, email, name string, roleID *int) {
	user := &domain.User{
		ID:       id,
		Email:    email,
		Name:     name,
		RoleID:   roleID,
		IsActive: true,
	}
	suite.db.Create(user)
}

func TestAuthIntegration_RegisterFlow(t *testing.T) {
	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	mahasiswaRole := suite.seedRole("mahasiswa")

	t.Run("CompleteRegisterFlow_Success", func(t *testing.T) {
		// Setup
		supabaseUserID := "supabase-user-uuid-123"
		suite.mockAuth.On("Register", mock.Anything, mock.MatchedBy(func(r domain.AuthServiceRegisterRequest) bool {
			return r.Email == "test@student.polije.ac.id" && r.Password == "password123" && !r.AutoConfirm
		})).Return(&domain.AuthServiceResponse{
			AccessToken:  "",
			RefreshToken: "",
			TokenType:    "bearer",
			ExpiresIn:    0,
			User: &domain.AuthServiceUserInfo{
				ID:    supabaseUserID,
				Email: "test@student.polije.ac.id",
				Name:  "Test User",
			},
		}, nil).Once()

		// Execute: Call usecase
		req := dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@student.polije.ac.id",
			Password: "password123",
		}

		result, err := suite.authUsecase.Register(context.Background(), req)

		// Verify: No error and correct response
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.NeedsConfirmation)
		assert.Contains(t, result.Message, "konfirmasi")

		// Verify: User was created in database
		savedUser, err := suite.userRepo.GetByEmail(context.Background(), "test@student.polije.ac.id")
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
		req := dto.RegisterRequest{
			Name:     "New User",
			Email:    "existing@student.polije.ac.id",
			Password: "password123",
		}

		result, err := suite.authUsecase.Register(context.Background(), req)

		// Verify: Should return conflict error
		require.Error(t, err)
		assert.Nil(t, result)

		var appErr *apperrors.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, fiber.StatusConflict, appErr.HTTPStatus)
	})

	t.Run("RegisterFlow_InvalidEmailDomain", func(t *testing.T) {
		// Execute
		req := dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@gmail.com", // Invalid domain
			Password: "password123",
		}

		result, err := suite.authUsecase.Register(context.Background(), req)

		// Verify: Should return validation error
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "polije.ac.id")
	})

	t.Run("RegisterFlow_TeacherEmailRejected", func(t *testing.T) {
		req := dto.RegisterRequest{
			Name:     "Teacher User",
			Email:    "teacher@teacher.polije.ac.id",
			Password: "password123",
		}

		result, err := suite.authUsecase.Register(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "mahasiswa")
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
		req := dto.AuthRequest{
			Email:    "login@student.polije.ac.id",
			Password: "correctpassword",
		}

		refreshToken, authResp, err := suite.authUsecase.Login(context.Background(), req)

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
		req := dto.AuthRequest{
			Email:    "user@student.polije.ac.id",
			Password: "wrongpassword",
		}

		refreshToken, authResp, err := suite.authUsecase.Login(context.Background(), req)

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
		req := dto.ResetPasswordRequest{
			Email: "reset@student.polije.ac.id",
		}
		err := suite.authUsecase.RequestPasswordReset(context.Background(), req)

		// Verify
		require.NoError(t, err)
		suite.mockAuth.AssertExpectations(t)
	})

	t.Run("RequestPasswordReset_UnknownEmailStillDelegates", func(t *testing.T) {
		suite.mockAuth.On("RequestPasswordReset", mock.Anything, "nonexistent@student.polije.ac.id", mock.Anything).Return(nil).Once()

		req := dto.ResetPasswordRequest{Email: "nonexistent@student.polije.ac.id"}
		err := suite.authUsecase.RequestPasswordReset(context.Background(), req)

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

		newRefresh, resp, err := suite.authUsecase.RefreshToken(context.Background(), "old_refresh")
		require.NoError(t, err)
		assert.Equal(t, "new_refresh", newRefresh)
		assert.Equal(t, "new_access", resp.AccessToken)
		assert.NotZero(t, resp.ExpiresAt)
	})

	t.Run("Logout_Success", func(t *testing.T) {
		suite.mockAuth.On("Logout", mock.Anything, "access_token").Return(nil).Once()
		err := suite.authUsecase.Logout(context.Background(), "access_token")
		require.NoError(t, err)
	})
}
