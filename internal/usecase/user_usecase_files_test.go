package usecase

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/storage"
)

func TestUserUsecase_BulkAssignRole_Success(t *testing.T) {
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

	roleID := uint(999)
	userIDs := []string{"user-1", "user-2"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	err := userUC.BulkAssignRole(context.Background(), userIDs, roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

func TestUserUsecase_BulkAssignRole_UserFetchError(t *testing.T) {
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

	result, err := userUC.GetUserPermissions(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User tidak ditemukan")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_GetProfile_InternalError(t *testing.T) {
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

	mockUserRepo.On("GetProfileWithCounts", mock.Anything, userID).Return(nil, 0, 0, errors.New("db error"))

	result, err := userUC.GetProfile(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Terjadi kesalahan pada server")

	mockUserRepo.AssertExpectations(t)
}

func TestUserUsecase_UpdateProfile_RepoUpdateError(t *testing.T) {
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
