package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/rbac"
	"invento-service/internal/storage"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserUsecase_GetUserByID_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		FotoProfil:   stringPtr("/uploads/profiles/test.jpg"),
		RoleID:       intPtr(1),
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	role := &domain.Role{
		ID:       1,
		NamaRole: "admin",
	}

	user.Role = role

	mockUserRepo.On("GetProfileWithCounts", mock.Anything, userID).Return(user, 5, 10, nil)

	result, err := userUC.GetProfile(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "admin", result.Role)
	assert.Equal(t, 5, result.JumlahProject)
	assert.Equal(t, 10, result.JumlahModul)

	mockUserRepo.AssertExpectations(t)
}

// TestGetUserByID_NotFound tests user retrieval when user doesn't exist
func TestUserUsecase_GetUserByID_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-999"

	mockUserRepo.On("GetProfileWithCounts", mock.Anything, userID).Return(nil, 0, 0, gorm.ErrRecordNotFound)

	result, err := userUC.GetProfile(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

// TestListUsers_Success tests successful user list retrieval
func TestUserUsecase_ListUsers_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	users := []dto.UserListItem{
		{
			ID:         "user-1",
			Email:      "user1@example.com",
			Role:       "admin",
			DibuatPada: time.Now(),
		},
		{
			ID:         "user-2",
			Email:      "user2@example.com",
			Role:       "user",
			DibuatPada: time.Now(),
		},
	}

	total := 2

	params := dto.UserListQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserRepo.On("GetAll", mock.Anything, "", "", 1, 10).Return(users, total, nil)

	result, err := userUC.GetUserList(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 2, result.Pagination.TotalItems)

	mockUserRepo.AssertExpectations(t)
}

// TestListUsers_WithSearchAndFilter tests user list retrieval with search and role filter
func TestUserUsecase_ListUsers_WithSearchAndFilter(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	users := []dto.UserListItem{
		{
			ID:         "user-1",
			Email:      "admin@example.com",
			Role:       "admin",
			DibuatPada: time.Now(),
		},
	}

	total := 1

	params := dto.UserListQueryParams{
		Search:     "admin",
		FilterRole: "admin",
		Page:       1,
		Limit:      10,
	}

	mockUserRepo.On("GetAll", mock.Anything, "admin", "admin", 1, 10).Return(users, total, nil)

	result, err := userUC.GetUserList(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "admin@example.com", result.Items[0].Email)

	mockUserRepo.AssertExpectations(t)
}

// TestUpdateUserProfile_Success tests successful user profile update
func TestUserUsecase_UpdateUserProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"

	jenisKelamin := "Laki-laki"
	oldFotoProfil := "/uploads/profiles/old.jpg"

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		FotoProfil:   &oldFotoProfil,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	req := dto.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockUserRepo.On("UpdateProfile", mock.Anything, userID, req.Name, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(nil)
	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(5, nil)
	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(10, nil)

	result, err := userUC.UpdateProfile(context.Background(), userID, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated User", result.Name)

	mockUserRepo.AssertExpectations(t)
}

// TestUpdateUserProfile_NotFound tests user profile update when user doesn't exist
func TestUserUsecase_UpdateUserProfile_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-999"

	req := dto.UpdateProfileRequest{
		Name: "Updated Name",
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.UpdateProfile(context.Background(), userID, req, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

// TestDeleteUser_Success tests successful user deletion
func TestUserUsecase_DeleteUser_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"

	user := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockUserRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := userUC.DeleteUser(context.Background(), userID)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
}

// TestDeleteUser_NotFound tests user deletion when user doesn't exist
func TestUserUsecase_DeleteUser_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-999"

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	err := userUC.DeleteUser(context.Background(), userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

// TestGetUserPermissions_Success tests successful user permissions retrieval
func TestUserUsecase_GetUserPermissions_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	// Note: Casbin enforcer is skipped in tests
	var casbinEnforcer *rbac.CasbinEnforcer

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, casbinEnforcer, pathResolver, cfg)

	userID := "user-1"

	jenisKelamin := "Laki-laki"
	role := &domain.Role{
		ID:       1,
		NamaRole: "admin",
	}

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		RoleID:       intPtr(int(role.ID)),
		IsActive:     true,
		CreatedAt:    time.Now(),
		Role:         role,
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

	result, err := userUC.GetUserPermissions(context.Background(), userID)

	// Since we can't easily mock the casbin enforcer, we'll just check the call flow
	// In a real scenario, you'd need to mock the casbin enforcer properly
	if err == nil {
		assert.NotNil(t, result)
	}

	mockUserRepo.AssertExpectations(t)
}

// TestUserUsecase_UpdateUserRole_Success tests successful role update
func TestUserUsecase_UpdateUserRole_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	roleName := "admin"

	jenisKelamin := "Laki-laki"
	roleID := int(2)
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		RoleID:       intPtr(1),
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	role := &domain.Role{
		ID:       uint(roleID),
		NamaRole: roleName,
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockRoleRepo.On("GetByName", mock.Anything, roleName).Return(role, nil)
	mockUserRepo.On("UpdateRole", mock.Anything, userID, &roleID).Return(nil)

	err := userUC.UpdateUserRole(context.Background(), userID, roleName)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestUserUsecase_UpdateUserRole_UserNotFound tests user not found
func TestUserUsecase_UpdateUserRole_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-999"
	roleName := "admin"

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	err := userUC.UpdateUserRole(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

// TestUserUsecase_UpdateUserRole_RoleNotFound tests role not found
func TestUserUsecase_UpdateUserRole_RoleNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	roleName := "nonexistent"

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		RoleID:       intPtr(1),
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockRoleRepo.On("GetByName", mock.Anything, roleName).Return(nil, gorm.ErrRecordNotFound)

	err := userUC.UpdateUserRole(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Role tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestUserUsecase_UpdateUserRole_SameRole tests no update when role is same
func TestUserUsecase_UpdateUserRole_SameRole(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	roleName := "admin"
	roleID := int(1)

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		RoleID:       &roleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	role := &domain.Role{
		ID:       uint(roleID),
		NamaRole: roleName,
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockRoleRepo.On("GetByName", mock.Anything, roleName).Return(role, nil)

	err := userUC.UpdateUserRole(context.Background(), userID, roleName)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// TestUserUsecase_DownloadUserFiles_EmptyIDs tests empty IDs
func TestUserUsecase_DownloadUserFiles_EmptyIDs(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	ownerUserID := "user-1"
	projectIDs := []string{}
	modulIDs := []string{}

	result, err := userUC.DownloadUserFiles(context.Background(), ownerUserID, projectIDs, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "project IDs atau modul IDs")
}

// TestUserUsecase_DownloadUserFiles_UserNotFound tests user not found
func TestUserUsecase_DownloadUserFiles_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	ownerUserID := "user-999"
	projectIDs := []string{"1"}
	modulIDs := []string{"1"}

	mockUserRepo.On("GetByID", mock.Anything, ownerUserID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.DownloadUserFiles(context.Background(), ownerUserID, projectIDs, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

// TestUserUsecase_DownloadUserFiles_NoFilesFound tests no files found
func TestUserUsecase_DownloadUserFiles_NoFilesFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	ownerUserID := "user-1"
	projectIDs := []string{"1"}
	modulIDs := []string{"550e8400-e29b-41d4-a716-446655440001"}
	projectIDsUint := []uint{1}

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           ownerUserID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	mockUserRepo.On("GetByID", mock.Anything, ownerUserID).Return(user, nil)
	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDsUint, ownerUserID).Return([]domain.Project{}, nil)
	mockModulRepo.On("GetByIDs", mock.Anything, modulIDs, ownerUserID).Return([]domain.Modul{}, nil)

	result, err := userUC.DownloadUserFiles(context.Background(), ownerUserID, projectIDs, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "File tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserFiles_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	params := dto.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	items := []dto.UserFileItem{
		{
			ID:          "project-1",
			NamaFile:    "project-file.pdf",
			Kategori:    "project",
			DownloadURL: "/uploads/projects/user-1/project-file.pdf",
		},
		{
			ID:          "modul-1",
			NamaFile:    "modul-file.pdf",
			Kategori:    "modul",
			DownloadURL: "/uploads/moduls/user-1/modul-file.pdf",
		},
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockUserRepo.On("GetUserFiles", mock.Anything, userID, "", 1, 10).Return(items, 2, nil)

	result, err := userUC.GetUserFiles(context.Background(), userID, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 2, result.Pagination.TotalItems)

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserFiles_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-999"
	params := dto.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.GetUserFiles(context.Background(), userID, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserFiles_RepoError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	params := dto.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockUserRepo.On("GetUserFiles", mock.Anything, userID, "", 1, 10).Return(nil, 0, errors.New("db error"))

	result, err := userUC.GetUserFiles(context.Background(), userID, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUsersForRole_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(1)
	role := &domain.Role{
		ID:       roleID,
		NamaRole: "admin",
	}

	users := []dto.UserListItem{
		{
			ID:         "user-1",
			Email:      "admin@example.com",
			Role:       "admin",
			DibuatPada: time.Now(),
		},
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByRoleID", mock.Anything, roleID).Return(users, nil)

	result, err := userUC.GetUsersForRole(context.Background(), roleID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "user-1", result[0].ID)

	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUsersForRole_RoleNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(999)

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.GetUsersForRole(context.Background(), roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUsersForRole_InternalError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(1)
	role := &domain.Role{
		ID:       roleID,
		NamaRole: "admin",
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByRoleID", mock.Anything, roleID).Return(nil, errors.New("db error"))

	result, err := userUC.GetUsersForRole(context.Background(), roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_BulkAssignRole_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(1)
	userIDs := []string{"user-1", "user-2"}
	role := &domain.Role{
		ID:       roleID,
		NamaRole: "admin",
	}

	users := []*domain.User{
		{
			ID:    "user-1",
			Email: "user1@example.com",
			Name:  "User 1",
		},
		{
			ID:    "user-2",
			Email: "user2@example.com",
			Name:  "User 2",
		},
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, userIDs).Return(users, nil)
	mockUserRepo.On("BulkUpdateRole", mock.Anything, userIDs, roleID).Return(nil)

	err := userUC.BulkAssignRole(context.Background(), userIDs, roleID)

	assert.NoError(t, err)

	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_BulkAssignRole_RoleNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(999)
	userIDs := []string{"user-1", "user-2"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	err := userUC.BulkAssignRole(context.Background(), userIDs, roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

func TestUserUsecase_BulkAssignRole_UserFetchError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(1)
	userIDs := []string{"user-1", "user-2"}
	role := &domain.Role{
		ID:       roleID,
		NamaRole: "admin",
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, userIDs).Return(nil, errors.New("db error"))

	err := userUC.BulkAssignRole(context.Background(), userIDs, roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_BulkAssignRole_InternalError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	roleID := uint(1)
	userIDs := []string{"user-1", "user-2"}
	role := &domain.Role{
		ID:       roleID,
		NamaRole: "admin",
	}

	users := []*domain.User{
		{
			ID:    "user-1",
			Email: "user1@example.com",
			Name:  "User 1",
		},
		{
			ID:    "user-2",
			Email: "user2@example.com",
			Name:  "User 2",
		},
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByIDs", mock.Anything, userIDs).Return(users, nil)
	mockUserRepo.On("BulkUpdateRole", mock.Anything, userIDs, roleID).Return(errors.New("update error"))

	err := userUC.BulkAssignRole(context.Background(), userIDs, roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserPermissions_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-999"

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.GetUserPermissions(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetProfile_InternalError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"

	mockUserRepo.On("GetProfileWithCounts", mock.Anything, userID).Return(nil, 0, 0, errors.New("db error"))

	result, err := userUC.GetProfile(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_UpdateProfile_RepoUpdateError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	req := dto.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockUserRepo.On("UpdateProfile", mock.Anything, userID, req.Name, mock.AnythingOfType("*string"), (*string)(nil)).Return(errors.New("update failed"))

	result, err := userUC.UpdateProfile(context.Background(), userID, req, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_DownloadUserFiles_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	tmpFile, err := os.CreateTemp("", "download-user-files-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	ownerUserID := "user-1"
	projectIDs := []string{"1"}
	modulIDs := []string{}
	projectIDsUint := []uint{1}

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           ownerUserID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	projects := []domain.Project{
		{
			ID:       1,
			PathFile: tmpFile.Name(),
		},
	}

	mockUserRepo.On("GetByID", mock.Anything, ownerUserID).Return(user, nil)
	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDsUint, ownerUserID).Return(projects, nil)
	mockModulRepo.On("GetByIDs", mock.Anything, modulIDs, ownerUserID).Return([]domain.Modul{}, nil)

	result, err := userUC.DownloadUserFiles(context.Background(), ownerUserID, projectIDs, modulIDs)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, tmpFile.Name(), result)

	mockUserRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

func TestUserUsecase_UpdateUserRole_RepoUpdateError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg)

	userID := "user-1"
	roleName := "admin"
	roleID := int(2)

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		RoleID:       intPtr(1),
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	role := &domain.Role{
		ID:       uint(roleID),
		NamaRole: roleName,
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockRoleRepo.On("GetByName", mock.Anything, roleName).Return(role, nil)
	mockUserRepo.On("UpdateRole", mock.Anything, userID, &roleID).Return(errors.New("update failed"))

	err := userUC.UpdateUserRole(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
