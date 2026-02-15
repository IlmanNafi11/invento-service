package usecase

import (
	goerrors "errors"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	mocks "invento-service/internal/usecase/test"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func assertRoleUsecaseAppError(t *testing.T, err error, expectedCode string, expectedStatus int, expectedMessage string) {
	t.Helper()
	assert.Error(t, err)
	var appErr *apperrors.AppError
	if assert.True(t, goerrors.As(err, &appErr), "error harus bertipe *AppError") {
		assert.Equal(t, expectedCode, appErr.Code)
		assert.Equal(t, expectedStatus, appErr.HTTPStatus)
		assert.Equal(t, expectedMessage, appErr.Message)
	}
}

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

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

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

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

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

	assertRoleUsecaseAppError(t, err, apperrors.ErrNotFound, fiber.StatusNotFound, "Role tidak ditemukan")
	assert.Nil(t, result)

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

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

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

	mockPermissionRepo.On("GetAllByResourceActions", req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
		{ID: 2, Resource: "projects", Action: "create", Label: "Create"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.MatchedBy(func(rolePermissions []domain.RolePermission) bool {
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

	assertRoleUsecaseAppError(t, err, apperrors.ErrConflict, fiber.StatusConflict, "Nama role sudah ada")
	assert.Nil(t, result)

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

	assertRoleUsecaseAppError(t, err, apperrors.ErrValidation, fiber.StatusBadRequest, "permission tidak boleh kosong")
	assert.Nil(t, result)
}

func TestRoleUsecase_CreateRole_CreateRepoError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := domain.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByName", "editor").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Create", mock.AnythingOfType("*domain.Role")).Return(assert.AnError)

	result, err := roleUC.CreateRole(req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_SetPermissionsError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := domain.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByName", "editor").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Create", mock.AnythingOfType("*domain.Role")).Run(func(args mock.Arguments) {
		role := args.Get(0).(*domain.Role)
		role.ID = 1
	}).Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", req.Permissions).Return(nil, assert.AnError)

	result, err := roleUC.CreateRole(req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
	mockPermissionRepo.AssertExpectations(t)
}

func TestRoleUsecase_CreateRole_SavePolicyError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	req := domain.RoleCreateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	mockRoleRepo.On("GetByName", "editor").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Create", mock.AnythingOfType("*domain.Role")).Run(func(args mock.Arguments) {
		role := args.Get(0).(*domain.Role)
		role.ID = 1
		role.CreatedAt = time.Now()
		role.UpdatedAt = time.Now()
	}).Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.AnythingOfType("[]domain.RolePermission")).Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(assert.AnError)

	result, err := roleUC.CreateRole(req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockCasbinEnforcer.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_UpdateRepoError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := domain.RoleUpdateRequest{
		NamaRole: "editor_updated",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}
	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", "editor_updated").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Update", mock.AnythingOfType("*domain.Role")).Return(assert.AnError)

	result, err := roleUC.UpdateRole(roleID, req)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleUsecase_UpdateRole_RemovePermissionsError(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermissionRepo := new(MockPermissionRepository)
	mockRolePermissionRepo := new(MockRolePermissionRepository)
	mockCasbinEnforcer := mocks.NewMockCasbinEnforcer()

	roleUC := NewRoleUsecase(mockRoleRepo, mockPermissionRepo, mockRolePermissionRepo, mockCasbinEnforcer)

	roleID := uint(1)
	req := domain.RoleUpdateRequest{
		NamaRole: "editor_updated",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor"}
	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRoleRepo.On("GetByName", "editor_updated").Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("Update", mock.AnythingOfType("*domain.Role")).Return(nil)
	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(assert.AnError)

	result, err := roleUC.UpdateRole(roleID, req)

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
	req := domain.RoleUpdateRequest{
		NamaRole:    "editor",
		Permissions: map[string][]string{},
	}

	result, err := roleUC.UpdateRole(roleID, req)

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
	req := domain.RoleUpdateRequest{
		NamaRole: "editor",
		Permissions: map[string][]string{
			"projects": {"read"},
		},
	}

	existingRole := &domain.Role{ID: 1, NamaRole: "editor", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	// No GetByName call since name is same
	mockRoleRepo.On("Update", mock.AnythingOfType("*domain.Role")).Return(nil)

	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)

	mockPermissionRepo.On("GetAllByResourceActions", req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.AnythingOfType("[]domain.RolePermission")).Return(nil)
	mockCasbinEnforcer.On("AddPermissionForRole", "editor", "projects", "read").Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(nil)

	result, err := roleUC.UpdateRole(roleID, req)

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

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(nil)
	mockRoleRepo.On("Delete", roleID).Return(assert.AnError)

	err := roleUC.DeleteRole(roleID)

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

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(nil)
	mockRoleRepo.On("Delete", roleID).Return(nil)
	mockCasbinEnforcer.On("SavePolicy").Return(assert.AnError)

	err := roleUC.DeleteRole(roleID)

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

	mockRoleRepo.On("GetByID", roleID).Return(role, nil)
	mockRolePermissionRepo.On("GetPermissionsForRole", roleID).Return(nil, assert.AnError)

	result, err := roleUC.GetRoleDetail(roleID)

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

	mockPermissionRepo.On("GetAllByResourceActions", req.Permissions).Return([]domain.Permission{
		{ID: 1, Resource: "projects", Action: "read", Label: "Read"},
		{ID: 3, Resource: "projects", Action: "update", Label: "Update"},
	}, nil)

	mockRolePermissionRepo.On("BulkCreate", mock.MatchedBy(func(rolePermissions []domain.RolePermission) bool {
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

	assertRoleUsecaseAppError(t, err, apperrors.ErrNotFound, fiber.StatusNotFound, "Role tidak ditemukan")

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

	mockRoleRepo.On("GetByID", roleID).Return(existingRole, nil)
	mockRolePermissionRepo.On("DeleteByRoleID", roleID).Return(nil)
	mockCasbinEnforcer.On("RemoveAllPermissionsForRole", "editor").Return(nil)
	mockCasbinEnforcer.On("DeleteRole", "editor").Return(assert.AnError)

	err := roleUC.DeleteRole(roleID)

	assertRoleUsecaseAppError(t, err, apperrors.ErrInternal, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")

	mockRoleRepo.AssertExpectations(t)
	mockRolePermissionRepo.AssertExpectations(t)
	mockCasbinEnforcer.AssertExpectations(t)
}
