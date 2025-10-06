package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"math"

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
}

func NewRoleUsecase(
	roleRepo repo.RoleRepository,
	permissionRepo repo.PermissionRepository,
	rolePermissionRepo repo.RolePermissionRepository,
	casbinEnforcer *helper.CasbinEnforcer,
) RoleUsecase {
	return &roleUsecase{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		casbinEnforcer:     casbinEnforcer,
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
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	roles, total, err := uc.roleRepo.GetAll(params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, errors.New("gagal mengambil daftar role")
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &domain.RoleListData{
		Items: roles,
		Pagination: domain.PaginationData{
			Page:       params.Page,
			Limit:      params.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (uc *roleUsecase) CreateRole(req domain.RoleCreateRequest) (*domain.RoleDetailResponse, error) {
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

	var permissionDetails []domain.RolePermissionDetail
	permissionCount := 0

	for resource, actions := range req.Permissions {
		var resourceActions []string
		for _, action := range actions {
			permission, err := uc.permissionRepo.GetByResourceAndAction(resource, action)
			if err != nil {
				continue
			}

			rolePermission := &domain.RolePermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
			}

			if err := uc.rolePermissionRepo.Create(rolePermission); err != nil {
				continue
			}

			if err := uc.casbinEnforcer.AddPermissionForRole(role.NamaRole, resource, action); err != nil {
				continue
			}

			resourceActions = append(resourceActions, action)
			permissionCount++
		}

		if len(resourceActions) > 0 {
			permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
				Resource: resource,
				Actions:  resourceActions,
			})
		}
	}

	if err := uc.casbinEnforcer.SavePolicy(); err != nil {
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

	resourceMap := make(map[string][]string)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], perm.Action)
	}

	var permissionDetails []domain.RolePermissionDetail
	for resource, actions := range resourceMap {
		permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
			Resource: resource,
			Actions:  actions,
		})
	}

	return &domain.RoleDetailResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		Permissions:      permissionDetails,
		JumlahPermission: len(permissions),
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}, nil
}

func (uc *roleUsecase) UpdateRole(id uint, req domain.RoleUpdateRequest) (*domain.RoleDetailResponse, error) {
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

	if err := uc.rolePermissionRepo.DeleteByRoleID(id); err != nil {
		return nil, errors.New("gagal menghapus permission lama")
	}

	if err := uc.casbinEnforcer.RemoveAllPermissionsForRole(oldRoleName); err != nil {
		return nil, errors.New("gagal menghapus policy casbin lama")
	}

	var permissionDetails []domain.RolePermissionDetail
	permissionCount := 0

	for resource, actions := range req.Permissions {
		var resourceActions []string
		for _, action := range actions {
			permission, err := uc.permissionRepo.GetByResourceAndAction(resource, action)
			if err != nil {
				continue
			}

			rolePermission := &domain.RolePermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
			}

			if err := uc.rolePermissionRepo.Create(rolePermission); err != nil {
				continue
			}

			if err := uc.casbinEnforcer.AddPermissionForRole(role.NamaRole, resource, action); err != nil {
				continue
			}

			resourceActions = append(resourceActions, action)
			permissionCount++
		}

		if len(resourceActions) > 0 {
			permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
				Resource: resource,
				Actions:  resourceActions,
			})
		}
	}

	if err := uc.casbinEnforcer.SavePolicy(); err != nil {
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

	if err := uc.rolePermissionRepo.DeleteByRoleID(id); err != nil {
		return errors.New("gagal menghapus permission role")
	}

	if err := uc.casbinEnforcer.DeleteRole(role.NamaRole); err != nil {
		return errors.New("gagal menghapus role dari casbin")
	}

	if err := uc.roleRepo.Delete(id); err != nil {
		return errors.New("gagal menghapus role")
	}

	if err := uc.casbinEnforcer.SavePolicy(); err != nil {
		return errors.New("gagal menyimpan policy casbin")
	}

	return nil
}
