package usecase

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	apperrors "invento-service/internal/errors"
	mocks "invento-service/internal/usecase/test"
)

func TestRoleUsecase_GetAvailablePermissions_Success(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	expectedPermissions := []dto.ResourcePermissions{
		{
			Name: "projects",
			Permissions: []dto.PermissionItem{
				{Action: "create", Label: "Create"},
				{Action: "read", Label: "Read"},
			},
		},
	}

	mockPermissionRepo.On("GetAvailablePermissions", mock.Anything).Return(expectedPermissions, nil)

	result, err := roleUC.GetAvailablePermissions(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "projects", result[0].Name)
	assert.Len(t, result[0].Permissions, 2)

	mockPermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetAvailablePermissions_Error(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	mockPermissionRepo.On("GetAvailablePermissions", mock.Anything).Return(nil, assert.AnError)

	result, err := roleUC.GetAvailablePermissions(context.Background())

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockPermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleList_Success(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roles := []dto.RoleListItem{
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

	mockRoleRepo.On("GetAll", mock.Anything, "", 1, 10).Return(roles, 2, nil)

	params := dto.RoleListQueryParams{
		Page:  1,
		Limit: 10,
	}

	result, err := roleUC.GetRoleList(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 2, result.Pagination.TotalItems)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleList_WithError(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	mockRoleRepo.On("GetAll", mock.Anything, "", 1, 10).Return(nil, 0, assert.AnError)

	params := dto.RoleListQueryParams{
		Page:  1,
		Limit: 10,
	}

	result, err := roleUC.GetRoleList(context.Background(), params)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleList_WithSearch(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roles := []dto.RoleListItem{
		{
			ID:                1,
			NamaRole:          "admin",
			JumlahPermission:  5,
			TanggalDiperbarui: time.Now(),
		},
	}

	mockRoleRepo.On("GetAll", mock.Anything, "admin", 1, 10).Return(roles, 1, nil)

	params := dto.RoleListQueryParams{
		Search: "admin",
		Page:   1,
		Limit:  10,
	}

	result, err := roleUC.GetRoleList(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "admin", result.Items[0].NamaRole)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleDetail_NotFound(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(999)

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := roleUC.GetRoleDetail(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrNotFound, fiber.StatusNotFound, "Role tidak ditemukan")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleDetail_DBError(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, nil)

	roleID := uint(1)

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	result, err := roleUC.GetRoleDetail(context.Background(), roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_GetRoleDetail_Success(t *testing.T) {
	t.Parallel()
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

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockRolePermissionRepo.On("GetPermissionsForRole", mock.Anything, roleID).Return(permissions, nil)

	result, err := roleUC.GetRoleDetail(context.Background(), roleID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "admin", result.NamaRole)
	assert.Equal(t, 2, result.JumlahPermission)

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_Success(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := dto.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read", "create"},
		},
	}

	mockRoleRepo.On("GetByName", mock.Anything, "editor").Return(nil, gorm.ErrRecordNotFound)

	mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Role")).Run(func(args mock.Arguments) {
		role := args.Get(1).(*domain.Role)
		role.ID = 1
		role.CreatedAt = time.Now()
		role.UpdatedAt = time.Now()
	}).Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", mock.Anything, req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
		{ID: 2, Resource: "projects", Action: "create", Label: "Create"},
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
		return ids[1] && ids[2]
	})).Return(nil)

	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "create").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	result, err := roleUC.CreateRole(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "editor", result.NamaRole)

	mockRoleRepo.AssertExpectations(t)
	mockPermissionRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_DuplicateName(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := dto.RoleCreateRequest{
		NamaRole: "admin",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "admin"}
	mockRoleRepo.On("GetByName", mock.Anything, "admin").Return(existingRole, nil)

	result, err := roleUC.CreateRole(context.Background(), req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrConflict, fiber.StatusConflict, "Nama role sudah ada")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_EmptyPermissions(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := dto.RoleCreateRequest{
		NamaRole:    "editor",
		Permissions: map[string][]string{},
	}

	result, err := roleUC.CreateRole(context.Background(), req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrValidation, fiber.StatusBadRequest, "permission tidak boleh kosong")
	assert.Nil(t, result)
}

func TestRoleUsecase_CreateRole_CreateRepoError(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := dto.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByName", mock.Anything, "editor").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Role")).Return(assert.AnError)

	result, err := roleUC.CreateRole(context.Background(), req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_SetPermissionsError(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := dto.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByName", mock.Anything, "editor").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Role")).Run(func(args mock.Arguments) {
		role := args.Get(1).(*domain.Role)
		role.ID = 1
	}).Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", mock.Anything, req.Permissions).Return(nil, assert.AnError)

	result, err := roleUC.CreateRole(context.Background(), req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
	mockPermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_SavePolicyError(t *testing.T) {
	t.Parallel()
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := dto.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByName", mock.Anything, "editor").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Role")).Run(func(args mock.Arguments) {
		role := args.Get(1).(*domain.Role)
		role.ID = 1
		role.CreatedAt = time.Now()
		role.UpdatedAt = time.Now()
	}).Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", mock.Anything, req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.Anything, mock.AnythingOfType("[]domain.RolePermission")).Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(assert.AnError)

	result, err := roleUC.CreateRole(context.Background(), req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_UpdateRepoError(t *testing.T) {
	t.Parallel()
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
	mockRoleRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Role")).Return(assert.AnError)

	result, err := roleUC.UpdateRole(context.Background(), roleID, req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}
