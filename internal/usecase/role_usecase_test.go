package usecase

import (
	"fiber-boiler-plate/internal/domain"
	mocks "fiber-boiler-plate/internal/usecase/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestRoleUsecase_GetAvailablePermissions_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

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

func TestRoleUsecase_GetRoleDetail_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(1)
	role := &domain.Role{
		ID:        1,
		NamaRole:  "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	permissions := []domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
		{ID: 2, Resource: "projects", Action: "create", Label: "Create"},
	}

	mockRoleRepo.On("GetByID", roleID).Return(role, nil)
	mockRolePermissionRepo.On("GetPermissionsForRole", roleID).Return(permissions, nil)

	result, err := roleUC.GetRoleDetail(roleID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "admin", result.NamaRole)
	assert.Equal(t, 2, result.JumlahPermission)

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := domain.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read", "create"},
		},
	}

	mockRoleRepo.On("GetByName", "editor").Return(nil, gorm.ErrRecordNotFound)

	mockRoleRepo.On("Create", mock.AnythingOfType("*domain.Role")).Run(func(args mock.Arguments) {
		role := args.Get(0).(*domain.Role)
		role.ID = 1
		role.CreatedAt = time.Now()
		role.UpdatedAt = time.Now()
	}).Return(nil)

	readPerm := &domain.Permission{ID: 1, Resource: "projects", Action: "read", Label: "Read"}
	createPerm := &domain.Permission{ID: 2, Resource: "projects", Action: "create", Label: "Create"}

	mockPermissionRepo.On("GetByResourceAndAction", "projects", "read").Return(readPerm, nil)
	mockPermissionRepo.On("GetByResourceAndAction", "projects", "create").Return(createPerm, nil)

	mockRolePermissionRepo.On("Create", mock.AnythingOfType("*domain.RolePermission")).Return(nil)

	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "create").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	result, err := roleUC.CreateRole(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "editor", result.NamaRole)

	mockRoleRepo.AssertExpectations(t)
	mockPermissionRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_DuplicateName(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := domain.RoleCreateRequest{
		NamaRole: "admin",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "admin"}
	mockRoleRepo.On("GetByName", "admin").Return(existingRole, nil)

	result, err := roleUC.CreateRole(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "nama role sudah ada")

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_EmptyPermissions(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := domain.RoleCreateRequest{
		NamaRole:    "editor",
		Permissions: map[string][]string{},
	}

	result, err := roleUC.CreateRole(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "permission tidak boleh kosong")
}

func TestRoleUsecase_UpdateRole_Success(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := domain.RoleUpdateRequest{
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

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", "editor_updated").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Update", mock.AnythingOfType("*domain.Role")).Return(nil)

	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)

	readPerm := &domain.Permission{ID: 1, Resource: "projects", Action: "read", Label: "Read"}
	updatePerm := &domain.Permission{ID: 3, Resource: "projects", Action: "update", Label: "Update"}

	mockPermissionRepo.On("GetByResourceAndAction", "projects", "read").Return(readPerm, nil)
	mockPermissionRepo.On("GetByResourceAndAction", "projects", "update").Return(updatePerm, nil)

	mockRolePermissionRepo.On("Create", mock.AnythingOfType("*domain.RolePermission")).Return(nil)

	mockCasbinEnforcer.On("AddPermissionForRole", "editor_updated", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor_updated", "projects", "update").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	result, err := roleUC.UpdateRole(roleID, req)

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
	req := domain.RoleUpdateRequest{
		NamaRole: "editor_updated",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByID", roleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := roleUC.UpdateRole(roleID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_DuplicateName(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := domain.RoleUpdateRequest{
		NamaRole: "admin",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}
	conflictingRole := &domain.Role{ID: 2, NamaRole: "admin"}

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", "admin").Return(conflictingRole, nil)

	result, err := roleUC.UpdateRole(roleID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "nama role sudah ada")

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

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(nil)
	mockRoleRepo.On("Delete", roleID).Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	err := roleUC.DeleteRole(roleID)

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

	mockRoleRepo.On("GetByID", roleID).Return(nil, gorm.ErrRecordNotFound)

	err := roleUC.DeleteRole(roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
}

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

func TestRoleUsecase_DeleteRole_CasbinDeleteRoleError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(assert.AnError)

	err := roleUC.DeleteRole(roleID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal menghapus role dari casbin")

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}
