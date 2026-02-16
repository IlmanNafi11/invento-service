package usecase

import (
	"context"
	"errors"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/rbac"
	"invento-service/internal/usecase/repo"

	"gorm.io/gorm"
)

type RoleUsecase interface {
	GetAvailablePermissions(ctx context.Context) ([]dto.ResourcePermissions, error)
	GetRoleList(ctx context.Context, params dto.RoleListQueryParams) (*dto.RoleListData, error)
	CreateRole(ctx context.Context, req dto.RoleCreateRequest) (*dto.RoleDetailResponse, error)
	GetRoleDetail(ctx context.Context, id uint) (*dto.RoleDetailResponse, error)
	UpdateRole(ctx context.Context, id uint, req dto.RoleUpdateRequest) (*dto.RoleDetailResponse, error)
	DeleteRole(ctx context.Context, id uint) error
}

type roleUsecase struct {
	roleRepo           repo.RoleRepository
	permissionRepo     repo.PermissionRepository
	rolePermissionRepo repo.RolePermissionRepository
	casbinEnforcer     rbac.CasbinEnforcerInterface
	rbacHelper         *rbac.RBACHelper
}

func NewRoleUsecase(
	roleRepo repo.RoleRepository,
	permissionRepo repo.PermissionRepository,
	rolePermissionRepo repo.RolePermissionRepository,
	casbinEnforcer rbac.CasbinEnforcerInterface,
) RoleUsecase {
	rbacHelper := rbac.NewRBACHelper(casbinEnforcer)

	return &roleUsecase{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		casbinEnforcer:     casbinEnforcer,
		rbacHelper:         rbacHelper,
	}
}

func (uc *roleUsecase) GetAvailablePermissions(ctx context.Context) ([]dto.ResourcePermissions, error) {
	permissions, err := uc.permissionRepo.GetAvailablePermissions(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}
	return permissions, nil
}

func (uc *roleUsecase) GetRoleList(ctx context.Context, params dto.RoleListQueryParams) (*dto.RoleListData, error) {
	paginationParams := httputil.NormalizePaginationParams(params.Page, params.Limit)

	roles, total, err := uc.roleRepo.GetAll(ctx, params.Search, paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	pagination := httputil.CalculatePagination(paginationParams.Page, paginationParams.Limit, total)

	return &dto.RoleListData{
		Items:      roles,
		Pagination: pagination,
	}, nil
}

func (uc *roleUsecase) CreateRole(ctx context.Context, req dto.RoleCreateRequest) (*dto.RoleDetailResponse, error) {
	if err := uc.rbacHelper.ValidatePermissionFormat(req.Permissions); err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	existingRole, _ := uc.roleRepo.GetByName(ctx, req.NamaRole)
	if existingRole != nil {
		return nil, apperrors.NewConflictError("Nama role sudah ada")
	}

	role := &domain.Role{
		NamaRole: req.NamaRole,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	permissionDetails, permissionCount, err := uc.rbacHelper.SetRolePermissions(
		ctx,
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

	return &dto.RoleDetailResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		Permissions:      permissionDetails,
		JumlahPermission: permissionCount,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}, nil
}

func (uc *roleUsecase) GetRoleDetail(ctx context.Context, id uint) (*dto.RoleDetailResponse, error) {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	permissions, err := uc.rolePermissionRepo.GetPermissionsForRole(ctx, id)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return uc.rbacHelper.BuildRoleDetailResponse(role, permissions), nil
}

func (uc *roleUsecase) UpdateRole(ctx context.Context, id uint, req dto.RoleUpdateRequest) (*dto.RoleDetailResponse, error) {
	if err := uc.rbacHelper.ValidatePermissionFormat(req.Permissions); err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	if role.NamaRole != req.NamaRole {
		existingRole, _ := uc.roleRepo.GetByName(ctx, req.NamaRole)
		if existingRole != nil {
			return nil, apperrors.NewConflictError("Nama role sudah ada")
		}
	}

	oldRoleName := role.NamaRole
	role.NamaRole = req.NamaRole

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	if err := uc.rbacHelper.RemoveAllRolePermissions(ctx, id, oldRoleName, uc.rolePermissionRepo); err != nil {
		return nil, apperrors.NewInternalError(errors.New(err.Error()))
	}

	permissionDetails, permissionCount, err := uc.rbacHelper.SetRolePermissions(
		ctx,
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

	return &dto.RoleDetailResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		Permissions:      permissionDetails,
		JumlahPermission: permissionCount,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}, nil
}

func (uc *roleUsecase) DeleteRole(ctx context.Context, id uint) error {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Role")
		}
		return apperrors.NewInternalError(err)
	}

	if err := uc.rbacHelper.RemoveAllRolePermissions(ctx, id, role.NamaRole, uc.rolePermissionRepo); err != nil {
		return apperrors.NewInternalError(errors.New(err.Error()))
	}

	if err := uc.casbinEnforcer.DeleteRole(role.NamaRole); err != nil {
		return apperrors.NewInternalError(err)
	}

	if err := uc.roleRepo.Delete(ctx, id); err != nil {
		return apperrors.NewInternalError(err)
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return apperrors.NewInternalError(err)
	}

	return nil
}
