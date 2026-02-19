package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/storage"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserUsecase_UpdateUserRole_RoleNotFound(t *testing.T) {
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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

	roleID := uint(999)

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userUC.GetUsersForRole(context.Background(), roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUsersForRole_InternalError(t *testing.T) {
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

	userUC := NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, nil, nil, pathResolver, cfg, zerolog.Nop())

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
