package usecase

import (
	"errors"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/helper"
	"invento-service/internal/usecase/repo"

	"gorm.io/gorm"
)

type RoleUsecase interface {
	GetAvailablePermissions() ([]domain.ResourcePermissions, error)
	GetRoleList(params domain.RoleListQueryParams) (*domain.RoleListData, error)
	CreateRole(req domain.RoleCreateRequest) (*domain.RoleDetailResponse, error)
	GetRoleDetail(id uint) (*domain.RoleDetailResponse, error)
	UpdateRole(id uint, req domain.RoleUpdateRequest) (*domain.RoleDetailResponse, error)
	DeleteRole(id uint) error
}

type roleUsecase struct {
	roleRepo           repo.RoleRepository
	permissionRepo     repo.PermissionRepository
	rolePermissionRepo repo.RolePermissionRepository
	casbinEnforcer     helper.CasbinEnforcerInterface
	rbacHelper         *helper.RBACHelper
}

func NewRoleUsecase(
	roleRepo repo.RoleRepository,
	permissionRepo repo.PermissionRepository,
	rolePermissionRepo repo.RolePermissionRepository,
	casbinEnforcer helper.CasbinEnforcerInterface,
) RoleUsecase {
	rbacHelper := helper.NewRBACHelper(casbinEnforcer)

	return &roleUsecase{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		casbinEnforcer:     casbinEnforcer,
		rbacHelper:         rbacHelper,
	}
}

func (uc *roleUsecase) GetAvailablePermissions() ([]domain.ResourcePermissions, error) {
	permissions, err := uc.permissionRepo.GetAvailablePermissions()
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}
	return permissions, nil
}

func (uc *roleUsecase) GetRoleList(params domain.RoleListQueryParams) (*domain.RoleListData, error) {
	paginationParams := helper.NormalizePaginationParams(params.Page, params.Limit)

	roles, total, err := uc.roleRepo.GetAll(params.Search, paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	pagination := helper.CalculatePagination(paginationParams.Page, paginationParams.Limit, total)

	return &domain.RoleListData{
		Items:      roles,
		Pagination: pagination,
	}, nil
}

func (uc *roleUsecase) CreateRole(req domain.RoleCreateRequest) (*domain.RoleDetailResponse, error) {
	if err := uc.rbacHelper.ValidatePermissionFormat(req.Permissions); err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	existingRole, _ := uc.roleRepo.GetByName(req.NamaRole)
	if existingRole != nil {
		return nil, apperrors.NewConflictError("Nama role sudah ada")
	}

	role := &domain.Role{
		NamaRole: req.NamaRole,
	}

	if err := uc.roleRepo.Create(role); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	permissionDetails, permissionCount, err := uc.rbacHelper.SetRolePermissions(
		role.ID,
		role.NamaRole,
		req.Permissions,
		uc.permissionRepo,
		uc.rolePermissionRepo,
	)
	if err != nil {
		return nil, apperrors.NewInternalError(errors.New(err.Error()))
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return &domain.RoleDetailResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		Permissions:      permissionDetails,
		JumlahPermission: permissionCount,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}, nil
}

func (uc *roleUsecase) GetRoleDetail(id uint) (*domain.RoleDetailResponse, error) {
	role, err := uc.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	permissions, err := uc.rolePermissionRepo.GetPermissionsForRole(id)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return uc.rbacHelper.BuildRoleDetailResponse(role, permissions), nil
}

func (uc *roleUsecase) UpdateRole(id uint, req domain.RoleUpdateRequest) (*domain.RoleDetailResponse, error) {
	if err := uc.rbacHelper.ValidatePermissionFormat(req.Permissions); err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	role, err := uc.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	if role.NamaRole != req.NamaRole {
		existingRole, _ := uc.roleRepo.GetByName(req.NamaRole)
		if existingRole != nil {
			return nil, apperrors.NewConflictError("Nama role sudah ada")
		}
	}

	oldRoleName := role.NamaRole
	role.NamaRole = req.NamaRole

	if err := uc.roleRepo.Update(role); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	if err := uc.rbacHelper.RemoveAllRolePermissions(id, oldRoleName, uc.rolePermissionRepo); err != nil {
		return nil, apperrors.NewInternalError(errors.New(err.Error()))
	}

	permissionDetails, permissionCount, err := uc.rbacHelper.SetRolePermissions(
		role.ID,
		role.NamaRole,
		req.Permissions,
		uc.permissionRepo,
		uc.rolePermissionRepo,
	)
	if err != nil {
		return nil, apperrors.NewInternalError(errors.New(err.Error()))
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return &domain.RoleDetailResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		Permissions:      permissionDetails,
		JumlahPermission: permissionCount,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}, nil
}

func (uc *roleUsecase) DeleteRole(id uint) error {
	role, err := uc.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Role")
		}
		return apperrors.NewInternalError(err)
	}

	if err := uc.rbacHelper.RemoveAllRolePermissions(id, role.NamaRole, uc.rolePermissionRepo); err != nil {
		return apperrors.NewInternalError(errors.New(err.Error()))
	}

	if err := uc.casbinEnforcer.DeleteRole(role.NamaRole); err != nil {
		return apperrors.NewInternalError(err)
	}

	if err := uc.roleRepo.Delete(id); err != nil {
		return apperrors.NewInternalError(err)
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return apperrors.NewInternalError(err)
	}

	return nil
}
