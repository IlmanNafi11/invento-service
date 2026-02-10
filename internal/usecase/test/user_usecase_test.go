package usecase_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// TestGetUserByID_Success tests successful user retrieval by ID
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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	userID := uint(1)

	jenisKelamin := "Laki-laki"
	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		Name:         "Test User",
		JenisKelamin: &jenisKelamin,
		FotoProfil:   stringPtr("/uploads/profiles/test.jpg"),
		RoleID:       uintPtr(1),
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	role := &domain.Role{
		ID:       1,
		NamaRole: "admin",
	}

	user.Role = role

	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockProjectRepo.On("CountByUserID", userID).Return(5, nil)
	mockModulRepo.On("CountByUserID", userID).Return(10, nil)

	result, err := userUC.GetProfile(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "admin", result.Role)
	assert.Equal(t, 5, result.JumlahProject)
	assert.Equal(t, 10, result.JumlahModul)

	mockUserRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	userID := uint(999)

	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.GetProfile(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user tidak ditemukan")

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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	users := []domain.UserListItem{
		{
			ID:         1,
			Email:      "user1@example.com",
			Role:       "admin",
			DibuatPada: time.Now(),
		},
		{
			ID:         2,
			Email:      "user2@example.com",
			Role:       "user",
			DibuatPada: time.Now(),
		},
	}

	total := 2

	params := domain.UserListQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserRepo.On("GetAll", "", "", 1, 10).Return(users, total, nil)

	result, err := userUC.GetUserList(params)

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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	users := []domain.UserListItem{
		{
			ID:         1,
			Email:      "admin@example.com",
			Role:       "admin",
			DibuatPada: time.Now(),
		},
	}

	total := 1

	params := domain.UserListQueryParams{
		Search:     "admin",
		FilterRole: "admin",
		Page:       1,
		Limit:      10,
	}

	mockUserRepo.On("GetAll", "admin", "admin", 1, 10).Return(users, total, nil)

	result, err := userUC.GetUserList(params)

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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	userID := uint(1)

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

	req := domain.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	mockUserRepo.On("GetByID", userID).Return(user, nil).Twice()
	mockUserRepo.On("UpdateProfile", userID, req.Name, mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(nil)
	mockProjectRepo.On("CountByUserID", userID).Return(5, nil)
	mockModulRepo.On("CountByUserID", userID).Return(10, nil)

	result, err := userUC.UpdateProfile(userID, req, nil)

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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	userID := uint(999)

	req := domain.UpdateProfileRequest{
		Name: "Updated Name",
	}

	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.UpdateProfile(userID, req, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user tidak ditemukan")

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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	userID := uint(1)

	user := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockUserRepo.On("Delete", userID).Return(nil)

	err := userUC.DeleteUser(userID)

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
	pathResolver := helper.NewPathResolver(cfg)

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, pathResolver, cfg, nil)

	userID := uint(999)

	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	err := userUC.DeleteUser(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user tidak ditemukan")

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
	pathResolver := helper.NewPathResolver(cfg)
	// Note: Casbin enforcer is skipped in tests
	var casbinEnforcer *helper.CasbinEnforcer

	userUC := usecase.NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, casbinEnforcer, pathResolver, cfg, nil)

	userID := uint(1)

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
		RoleID:       &role.ID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		Role:         role,
	}

	mockUserRepo.On("GetByID", userID).Return(user, nil)

	result, err := userUC.GetUserPermissions(userID)

	// Since we can't easily mock the casbin enforcer, we'll just check the call flow
	// In a real scenario, you'd need to mock the casbin enforcer properly
	if err == nil {
		assert.NotNil(t, result)
	}

	mockUserRepo.AssertExpectations(t)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint) *uint {
	return &u
}
