package usecase

import (
	"fiber-boiler-plate/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestRoleUsecase_GetAvailablePermissions_Success tests successful permission retrieval
func TestRoleUsecase_GetAvailablePermissions_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	// Use nil for casbin enforcer - the usecase should handle it
	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	expectedPermissions := []domain.ResourcePermissions{
		{
			Name: "projects",
			Permissions: []domain.PermissionItem{
				{Action: "create", Label: "Create"},
				{Action: "read", Label: "Read"},
			},
		},
	}

	mockPermissionRepo.On("GetAvailablePermissions").Return(expectedPermissions, nil)

	result, err := roleUC.GetAvailablePermissions()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "projects", result[0].Name)
	assert.Len(t, result[0].Permissions, 2)

	mockPermissionRepo.AssertExpectations(t)
}

// TestRoleUsecase_GetAvailablePermissions_Error tests error handling
func TestRoleUsecase_GetAvailablePermissions_Error(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	mockPermissionRepo.On("GetAvailablePermissions").Return(nil, assert.AnError)

	result, err := roleUC.GetAvailablePermissions()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "gagal mengambil daftar permission")

	mockPermissionRepo.AssertExpectations(t)
}

// TestRoleUsecase_GetRoleList_Success tests successful role list retrieval
func TestRoleUsecase_GetRoleList_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roles := []domain.RoleListItem{
		{
			ID:                1,
			NamaRole:          "admin",
			JumlahPermission:  5,
			TanggalDiperbarui: time.Now(),
		},
		{
			ID:                2,
			NamaRole:          "user",
			JumlahPermission:  2,
			TanggalDiperbarui: time.Now(),
		},
	}

	mockRoleRepo.On("GetAll", "", 1, 10).Return(roles, 2, nil)

	params := domain.RoleListQueryParams{
		Page:  1,
		Limit: 10,
	}

	result, err := roleUC.GetRoleList(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 2, result.Pagination.TotalItems)

	mockRoleRepo.AssertExpectations(t)
}

// TestRoleUsecase_GetRoleList_WithError tests error handling
func TestRoleUsecase_GetRoleList_WithError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	mockRoleRepo.On("GetAll", "", 1, 10).Return(nil, 0, assert.AnError)

	params := domain.RoleListQueryParams{
		Page:  1,
		Limit: 10,
	}

	result, err := roleUC.GetRoleList(params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "gagal mengambil daftar role")

	mockRoleRepo.AssertExpectations(t)
}

// TestRoleUsecase_GetRoleList_WithSearch tests role list with search parameter
func TestRoleUsecase_GetRoleList_WithSearch(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roles := []domain.RoleListItem{
		{
			ID:                1,
			NamaRole:          "admin",
			JumlahPermission:  5,
			TanggalDiperbarui: time.Now(),
		},
	}

	mockRoleRepo.On("GetAll", "admin", 1, 10).Return(roles, 1, nil)

	params := domain.RoleListQueryParams{
		Search: "admin",
		Page:   1,
		Limit:  10,
	}

	result, err := roleUC.GetRoleList(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "admin", result.Items[0].NamaRole)

	mockRoleRepo.AssertExpectations(t)
}

// TestRoleUsecase_GetRoleDetail_NotFound tests role detail when role not found
func TestRoleUsecase_GetRoleDetail_NotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(999)

	mockRoleRepo.On("GetByID", roleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := roleUC.GetRoleDetail(roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

// TestRoleUsecase_GetRoleDetail_DBError tests database error handling
func TestRoleUsecase_GetRoleDetail_DBError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(1)

	mockRoleRepo.On("GetByID", roleID).Return(nil, assert.AnError)

	result, err := roleUC.GetRoleDetail(roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "gagal mengambil detail role")

	mockRoleRepo.AssertExpectations(t)
}

// TestRoleUsecase_DeleteRole_NotFound tests role deletion when role not found
func TestRoleUsecase_DeleteRole_NotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(999)

	mockRoleRepo.On("GetByID", roleID).Return(nil, gorm.ErrRecordNotFound)

	err := roleUC.DeleteRole(roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

// TestRoleUsecase_DeleteRole_DBError tests database error during deletion
func TestRoleUsecase_DeleteRole_DBError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(1)

	mockRoleRepo.On("GetByID", roleID).Return(nil, assert.AnError)

	err := roleUC.DeleteRole(roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal mengambil role")

	mockRoleRepo.AssertExpectations(t)
}
