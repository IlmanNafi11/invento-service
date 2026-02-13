package usecase

import (
	"context"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyJWT(token string) (domain.AuthClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.AuthClaims), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, req domain.AuthServiceRegisterRequest) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthServiceResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthServiceResponse), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockAuthService) RequestPasswordReset(ctx context.Context, email string, redirectTo string) error {
	args := m.Called(ctx, email, redirectTo)
	return args.Error(0)
}

func (m *MockAuthService) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

// MockUserRepository is a mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id string) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error {
	args := m.Called(userID, name, jenisKelamin, fotoProfil)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	args := m.Called(search, filterRole, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.UserListItem), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) UpdateRole(userID string, roleID *int) error {
	args := m.Called(userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetByRoleID(roleID uint) ([]domain.UserListItem, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserListItem), args.Error(1)
}

func (m *MockUserRepository) BulkUpdateRole(userIDs []string, roleID uint) error {
	args := m.Called(userIDs, roleID)
	return args.Error(0)
}

// MockRoleRepository is a mock for RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(id uint) (*domain.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(name string) (*domain.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetAll(search string, page, limit int) ([]domain.RoleListItem, int, error) {
	args := m.Called(search, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.RoleListItem), args.Int(1), args.Error(2)
}

// MockProjectRepository is a mock for ProjectRepository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(project *domain.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByUserID(userID string, search string, filterSemester int, filterKategori string, page, limit int) ([]domain.ProjectListItem, int, error) {
	args := m.Called(userID, search, filterSemester, filterKategori, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.ProjectListItem), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) GetByID(id uint) (*domain.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(project *domain.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByIDs(projectIDs []uint, userID string) ([]domain.Project, error) {
	args := m.Called(projectIDs, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *MockProjectRepository) CountByUserID(userID string) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockProjectRepository) GetByIDsForUser(projectIDs []uint, userID string) ([]domain.Project, error) {
	args := m.Called(projectIDs, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

// MockModulRepository is a mock for ModulRepository
type MockModulRepository struct {
	mock.Mock
}

func (m *MockModulRepository) Create(modul *domain.Modul) error {
	args := m.Called(modul)
	return args.Error(0)
}

func (m *MockModulRepository) GetByID(id string) (*domain.Modul, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByIDs(ids []string, userID string) ([]domain.Modul, error) {
	args := m.Called(ids, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByIDsForUser(ids []string, ownerUserID string) ([]domain.Modul, error) {
	args := m.Called(ids, ownerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByUserID(userID string, search string, filterType string, filterStatus string, page, limit int) ([]domain.ModulListItem, int, error) {
	args := m.Called(userID, search, filterType, filterStatus, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.ModulListItem), args.Int(1), args.Error(2)
}

func (m *MockModulRepository) CountByUserID(userID string) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockModulRepository) Update(modul *domain.Modul) error {
	args := m.Called(modul)
	return args.Error(0)
}

func (m *MockModulRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockModulRepository) UpdateMetadata(modul *domain.Modul) error {
	args := m.Called(modul)
	return args.Error(0)
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint) *uint {
	return &u
}

func intPtr(i int) *int {
	return &i
}

// MockTusModulUploadRepository is a mock for TusModulUploadRepository
type MockTusModulUploadRepository struct {
	mock.Mock
}

func (m *MockTusModulUploadRepository) Create(upload *domain.TusModulUpload) error {
	args := m.Called(upload)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) GetByID(id string) (*domain.TusModulUpload, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetByUserID(userID string) ([]domain.TusModulUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) UpdateOffset(id string, offset int64, progress float64) error {
	args := m.Called(id, offset, progress)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Complete(id string, modulID string, filePath string) error {
	args := m.Called(id, modulID, filePath)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) GetExpiredUploads(before time.Time) ([]domain.TusModulUpload, error) {
	args := m.Called(before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error) {
	args := m.Called(timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) CountActiveByUserID(userID string) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveByUserID(userID string) ([]domain.TusModulUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveUploadIDs() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockTusUploadRepository is a mock for TusUploadRepository
type MockTusUploadRepository struct {
	mock.Mock
}

func (m *MockTusUploadRepository) Create(upload *domain.TusUpload) error {
	args := m.Called(upload)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetByID(id string) (*domain.TusUpload, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetByUserID(userID string) ([]domain.TusUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) UpdateOffset(id string, offset int64, progress float64) error {
	args := m.Called(id, offset, progress)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetExpiredUploads(before time.Time) ([]domain.TusUpload, error) {
	args := m.Called(before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusUpload, error) {
	args := m.Called(timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetActiveByUserID(userID string) ([]domain.TusUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) Complete(id string, projectID uint, filePath string) error {
	args := m.Called(id, projectID, filePath)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetByUserIDAndStatus(userID string, status string) ([]domain.TusUpload, error) {
	args := m.Called(userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTusUploadRepository) ListActive() ([]domain.TusUpload, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) CountActiveByUserID(userID string) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusUploadRepository) GetActiveUploadIDs() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// Helper functions for test configs
func getTestModulConfig() *config.Config {
	return &config.Config{
		Upload: config.UploadConfig{
			MaxSize:              524288000, // 500 MB
			MaxSizeModul:         52428800,  // 50 MB
			MaxQueueModulPerUser: 5,
			IdleTimeout:          600, // 10 minutes
		},
	}
}

func getTestTusUploadConfig() *config.Config {
	return &config.Config{
		Upload: config.UploadConfig{
			MaxSize:              524288000, // 500 MB
			MaxSizeProject:       524288000, // 500 MB
			MaxConcurrentProject: 1,
			IdleTimeout:          600, // 10 minutes
		},
	}
}

// MockPermissionRepository is a mock for PermissionRepository
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(permission *domain.Permission) error {
	args := m.Called(permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetByID(id uint) (*domain.Permission, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetByResourceAndAction(resource, action string) (*domain.Permission, error) {
	args := m.Called(resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetAllByResourceActions(permissions map[string][]string) ([]domain.Permission, error) {
	args := m.Called(permissions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetAll() ([]domain.Permission, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetAvailablePermissions() ([]domain.ResourcePermissions, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ResourcePermissions), args.Error(1)
}

func (m *MockPermissionRepository) BulkCreate(permissions []domain.Permission) error {
	args := m.Called(permissions)
	return args.Error(0)
}

// MockRolePermissionRepository is a mock for RolePermissionRepository
type MockRolePermissionRepository struct {
	mock.Mock
}

func (m *MockRolePermissionRepository) Create(rolePermission *domain.RolePermission) error {
	args := m.Called(rolePermission)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) BulkCreate(rolePermissions []domain.RolePermission) error {
	args := m.Called(rolePermissions)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) GetByRoleID(roleID uint) ([]domain.RolePermission, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) DeleteByRoleID(roleID uint) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) GetPermissionsForRole(roleID uint) ([]domain.Permission, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}
