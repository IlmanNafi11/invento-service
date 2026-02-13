package repo

import (
	"time"

	"fiber-boiler-plate/internal/domain"
)

type UserRepository interface {
	GetByEmail(email string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
	Create(user *domain.User) error
	GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error)
	UpdateRole(userID string, roleID *int) error
	UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error
	Delete(userID string) error
	GetByRoleID(roleID uint) ([]domain.UserListItem, error)
	BulkUpdateRole(userIDs []string, roleID uint) error
}

type RoleRepository interface {
	Create(role *domain.Role) error
	GetByID(id uint) (*domain.Role, error)
	GetByName(name string) (*domain.Role, error)
	Update(role *domain.Role) error
	Delete(id uint) error
	GetAll(search string, page, limit int) ([]domain.RoleListItem, int, error)
}

type PermissionRepository interface {
	Create(permission *domain.Permission) error
	GetByID(id uint) (*domain.Permission, error)
	GetByResourceAndAction(resource, action string) (*domain.Permission, error)
	GetAll() ([]domain.Permission, error)
	GetAvailablePermissions() ([]domain.ResourcePermissions, error)
	BulkCreate(permissions []domain.Permission) error
}

type RolePermissionRepository interface {
	Create(rolePermission *domain.RolePermission) error
	GetByRoleID(roleID uint) ([]domain.RolePermission, error)
	DeleteByRoleID(roleID uint) error
	GetPermissionsForRole(roleID uint) ([]domain.Permission, error)
}

type ProjectRepository interface {
	Create(project *domain.Project) error
	GetByID(id uint) (*domain.Project, error)
	GetByIDs(ids []uint, userID string) ([]domain.Project, error)
	GetByIDsForUser(ids []uint, ownerUserID string) ([]domain.Project, error)
	GetByUserID(userID string, search string, filterSemester int, filterKategori string, page, limit int) ([]domain.ProjectListItem, int, error)
	CountByUserID(userID string) (int, error)
	Update(project *domain.Project) error
	Delete(id uint) error
}

type ModulRepository interface {
	Create(modul *domain.Modul) error
	GetByID(id string) (*domain.Modul, error)
	GetByIDs(ids []string, userID string) ([]domain.Modul, error)
	GetByIDsForUser(ids []string, ownerUserID string) ([]domain.Modul, error)
	GetByUserID(userID string, search string, filterType string, filterStatus string, page, limit int) ([]domain.ModulListItem, int, error)
	CountByUserID(userID string) (int, error)
	Update(modul *domain.Modul) error
	Delete(id string) error
	UpdateMetadata(modul *domain.Modul) error
}

type TusUploadRepository interface {
	Create(upload *domain.TusUpload) error
	GetByID(id string) (*domain.TusUpload, error)
	GetByUserID(userID string) ([]domain.TusUpload, error)
	CountActiveByUserID(userID string) (int64, error)
	UpdateOffset(id string, offset int64, progress float64) error
	UpdateStatus(id string, status string) error
	GetExpired(before time.Time) ([]domain.TusUpload, error)
	GetByUserIDAndStatus(userID string, status string) ([]domain.TusUpload, error)
	Delete(id string) error
	ListActive() ([]domain.TusUpload, error)
	GetActiveUploadIDs() ([]string, error)
	UpdateOffsetOnly(id string, offset int64) error
	UpdateUpload(upload *domain.TusUpload) error
}

type TusModulUploadRepository interface {
	Create(upload *domain.TusModulUpload) error
	GetByID(id string) (*domain.TusModulUpload, error)
	GetByUserID(userID string) ([]domain.TusModulUpload, error)
	UpdateOffset(id string, offset int64, progress float64) error
	UpdateStatus(id string, status string) error
	Complete(id string, modulID string, filePath string) error
	Delete(id string) error
	GetExpiredUploads() ([]domain.TusModulUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error)
	CountActiveByUserID(userID string) (int, error)
	GetActiveByUserID(userID string) ([]domain.TusModulUpload, error)
	GetActiveUploadIDs() ([]string, error)
}
