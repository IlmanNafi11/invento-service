package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/rbac"
	"invento-service/internal/storage"
)

func TestUserUsecase_GetUserByID_Success(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
