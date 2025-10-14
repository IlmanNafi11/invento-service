package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

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
	casbinEnforcer     *helper.CasbinEnforcer
	rbacHelper         *helper.RBACHelper
}

func NewRoleUsecase(
	roleRepo repo.RoleRepository,
	permissionRepo repo.PermissionRepository,
	rolePermissionRepo repo.RolePermissionRepository,
	casbinEnforcer *helper.CasbinEnforcer,
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
		return nil, errors.New("gagal mengambil daftar permission")
	}
	return permissions, nil
}

func (uc *roleUsecase) GetRoleList(params domain.RoleListQueryParams) (*domain.RoleListData, error) {
	paginationParams := helper.NormalizePaginationParams(params.Page, params.Limit)

	roles, total, err := uc.roleRepo.GetAll(params.Search, paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return nil, errors.New("gagal mengambil daftar role")
	}

	pagination := helper.CalculatePagination(paginationParams.Page, paginationParams.Limit, total)

	return &domain.RoleListData{
		Items:      roles,
		Pagination: pagination,
	}, nil
}

func (uc *roleUsecase) CreateRole(req domain.RoleCreateRequest) (*domain.RoleDetailResponse, error) {
	if err := uc.rbacHelper.ValidatePermissionFormat(req.Permissions); err != nil {
		return nil, err
	}

	existingRole, _ := uc.roleRepo.GetByName(req.NamaRole)
	if existingRole != nil {
		return nil, errors.New("nama role sudah ada")
	}

	role := &domain.Role{
		NamaRole: req.NamaRole,
	}

	if err := uc.roleRepo.Create(role); err != nil {
		return nil, errors.New("gagal membuat role")
	}

	permissionDetails, permissionCount, err := uc.rbacHelper.CreateRolePermissions(
		role.ID,
		req.Permissions,
		uc.permissionRepo,
		uc.rolePermissionRepo,
	)
	if err != nil {
		return nil, err
	}

	_, _, err = uc.rbacHelper.SyncPermissionsToRole(
		role.NamaRole,
		req.Permissions,
		uc.permissionRepo,
	)
	if err != nil {
		return nil, errors.New("gagal sync permissions ke casbin")
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return nil, errors.New("gagal menyimpan policy casbin")
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
			return nil, errors.New("role tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil detail role")
	}

	permissions, err := uc.rolePermissionRepo.GetPermissionsForRole(id)
	if err != nil {
		return nil, errors.New("gagal mengambil permission role")
	}

	return uc.rbacHelper.BuildRoleDetailResponse(role, permissions), nil
}

func (uc *roleUsecase) UpdateRole(id uint, req domain.RoleUpdateRequest) (*domain.RoleDetailResponse, error) {
	if err := uc.rbacHelper.ValidatePermissionFormat(req.Permissions); err != nil {
		return nil, err
	}

	role, err := uc.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil role")
	}

	if role.NamaRole != req.NamaRole {
		existingRole, _ := uc.roleRepo.GetByName(req.NamaRole)
		if existingRole != nil {
			return nil, errors.New("nama role sudah ada")
		}
	}

	oldRoleName := role.NamaRole
	role.NamaRole = req.NamaRole

	if err := uc.roleRepo.Update(role); err != nil {
		return nil, errors.New("gagal memperbarui role")
	}

	if err := uc.rbacHelper.RemoveAllRolePermissions(id, oldRoleName, uc.rolePermissionRepo); err != nil {
		return nil, err
	}

	permissionDetails, permissionCount, err := uc.rbacHelper.CreateRolePermissions(
		role.ID,
		req.Permissions,
		uc.permissionRepo,
		uc.rolePermissionRepo,
	)
	if err != nil {
		return nil, err
	}

	_, _, err = uc.rbacHelper.SyncPermissionsToRole(
		role.NamaRole,
		req.Permissions,
		uc.permissionRepo,
	)
	if err != nil {
		return nil, errors.New("gagal sync permissions ke casbin")
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return nil, errors.New("gagal menyimpan policy casbin")
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
			return errors.New("role tidak ditemukan")
		}
		return errors.New("gagal mengambil role")
	}

	if err := uc.rbacHelper.RemoveAllRolePermissions(id, role.NamaRole, uc.rolePermissionRepo); err != nil {
		return err
	}

	if err := uc.casbinEnforcer.DeleteRole(role.NamaRole); err != nil {
		return errors.New("gagal menghapus role dari casbin")
	}

	if err := uc.roleRepo.Delete(id); err != nil {
		return errors.New("gagal menghapus role")
	}

	if err := uc.rbacHelper.SavePolicy(); err != nil {
		return errors.New("gagal menyimpan policy casbin")
	}

	return nil
}
