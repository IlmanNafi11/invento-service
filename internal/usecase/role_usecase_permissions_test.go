package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	mocks "invento-service/internal/usecase/test"
)

func TestRoleUsecase_UpdateRole_RemovePermissionsError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := dto.RoleUpdateRequest{
		NamaRole: "editor_updated",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", mock.Anything, "editor_updated").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Role")).Return(nil)
	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(assert.AnError)

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_EmptyPermissions(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := dto.RoleUpdateRequest{
		NamaRole:    "editor",
		Permissions: map[string][]string{},
	}

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrValidation, fiber.StatusBadRequest, "permission tidak boleh kosong")
	assert.Nil(t, result)
}

func TestRoleUsecase_UpdateRole_SameNameNoConflict(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := dto.RoleUpdateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	// No GetByName call since name is same
	mockRoleRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Role")).Return(nil)

	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", mock.Anything, req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.Anything, mock.AnythingOfType("[]domain.RolePermission")).Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "editor", result.NamaRole)

	mockRoleRepo.AssertExpectations(t)
	mockPermissionRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_DeleteRole_DeleteRepoError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(nil)
	mockRoleRepo.On("Delete", mock.Anything, roleID).Return(assert.AnError)

	err := roleUC.DeleteRole(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_DeleteRole_SavePolicyError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(nil)
	mockRoleRepo.On("Delete", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(assert.AnError)

	err := roleUC.DeleteRole(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleDetail_PermissionRepoError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(1)
	role := &domain.Role{ID: 1, NamaRole: "admin"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockRolePermissionRepo.On("GetPermissionsForRole", mock.Anything, roleID).Return(nil, assert.AnError)

	result, err := roleUC.GetRoleDetail(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := dto.RoleUpdateRequest{
		NamaRole: "editor_updated",
		Permissions: map[string][]string{
			"projects": {"read", "update"},
		},
	}

	existingRole := &domain.Role{
		ID:        1,
		NamaRole:  "editor",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", mock.Anything, "editor_updated").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Role")).Return(nil)

	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", mock.Anything, req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
		{ID: 3, Resource: "projects", Action: "update", Label: "Update"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.Anything, mock.MatchedBy(func(rolePermissions []domain.RolePermission) bool {
		if len(rolePermissions) != 2 {
			return false
		}
		ids := map[uint]bool{}
		for _, rp := range rolePermissions {
			if rp.RoleID != 1 {
				return false
			}
			ids[rp.PermissionID] = true
		}
		return ids[1] && ids[3]
	})).Return(nil)

	mockCasbinEnforcer.On("AddPermissionForRole", "editor_updated", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor_updated", "projects", "update").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "editor_updated", result.NamaRole)

	mockRoleRepo.AssertExpectations(t)
	mockPermissionRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_NotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(999)
	req := dto.RoleUpdateRequest{
		NamaRole: "editor_updated",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrNotFound, fiber.StatusNotFound, "Role tidak ditemukan")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_DuplicateName(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := dto.RoleUpdateRequest{
		NamaRole: "admin",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}
	conflictingRole := &domain.Role{ID: 2, NamaRole: "admin"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", mock.Anything, "admin").Return(conflictingRole, nil)

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrConflict, fiber.StatusConflict, "Nama role sudah ada")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_DeleteRole_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(nil)
	mockRoleRepo.On("Delete", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	err := roleUC.DeleteRole(context.Background(), roleID)

	assert.NoError(t, err)

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_DeleteRole_NotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(999)

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	err := roleUC.DeleteRole(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrNotFound, fiber.StatusNotFound, "Role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_DeleteRole_DBError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(1)

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	err := roleUC.DeleteRole(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_DeleteRole_CasbinDeleteRoleError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", mock.Anything, roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(assert.AnError)

	err := roleUC.DeleteRole(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}
