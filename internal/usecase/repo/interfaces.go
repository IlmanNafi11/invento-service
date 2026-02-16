package repo

import (
	"context"
	"time"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
)

type UserRepository interface {
	GetByEmail(email string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
	GetByIDs(userIDs []string) ([]*domain.User, error)
	Create(user *domain.User) error
	GetAll(search, filterRole string, page, limit int) ([]dto.UserListItem, int, error)
	GetProfileWithCounts(userID string) (*domain.User, int, int, error)
	GetUserFiles(userID string, search string, page, limit int) ([]dto.UserFileItem, int, error)
	UpdateRole(userID string, roleID *int) error
	UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error
	Delete(userID string) error
	GetByRoleID(roleID uint) ([]dto.UserListItem, error)
	BulkUpdateRole(userIDs []string, roleID uint) error
}

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id uint) (*domain.Role, error)
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context, search string, page, limit int) ([]dto.RoleListItem, int, error)
}

type PermissionRepository interface {
	Create(ctx context.Context, permission *domain.Permission) error
	GetByID(ctx context.Context, id uint) (*domain.Permission, error)
	GetByResourceAndAction(ctx context.Context, resource, action string) (*domain.Permission, error)
	GetAllByResourceActions(ctx context.Context, permissions map[string][]string) ([]domain.Permission, error)
	GetAll(ctx context.Context) ([]domain.Permission, error)
	GetAvailablePermissions(ctx context.Context) ([]dto.ResourcePermissions, error)
	BulkCreate(ctx context.Context, permissions []domain.Permission) error
}

type RolePermissionRepository interface {
	Create(ctx context.Context, rolePermission *domain.RolePermission) error
	BulkCreate(ctx context.Context, rolePermissions []domain.RolePermission) error
	GetByRoleID(ctx context.Context, roleID uint) ([]domain.RolePermission, error)
	DeleteByRoleID(ctx context.Context, roleID uint) error
	GetPermissionsForRole(ctx context.Context, roleID uint) ([]domain.Permission, error)
}

type ProjectRepository interface {
	Create(project *domain.Project) error
	GetByID(id uint) (*domain.Project, error)
	GetByIDs(ids []uint, userID string) ([]domain.Project, error)
	GetByUserID(userID string, search string, filterSemester int, filterKategori string, page, limit int) ([]dto.ProjectListItem, int, error)
	CountByUserID(userID string) (int, error)
	Update(project *domain.Project) error
	Delete(id uint) error
}

type ModulRepository interface {
	Create(modul *domain.Modul) error
	GetByID(id string) (*domain.Modul, error)
	GetByIDs(ids []string, userID string) ([]domain.Modul, error)
	GetByUserID(userID string, search string, filterType string, filterStatus string, page, limit int) ([]dto.ModulListItem, int, error)
	CountByUserID(userID string) (int, error)
	Update(modul *domain.Modul) error
	Delete(id string) error
	UpdateMetadata(modul *domain.Modul) error
}

type TusUploadRepository interface {
	Create(upload *domain.TusUpload) error
	GetByID(id string) (*domain.TusUpload, error)
	GetByUserID(userID string) ([]domain.TusUpload, error)
	GetActiveByUserID(userID string) ([]domain.TusUpload, error)
	CountActiveByUserID(userID string) (int64, error)
	UpdateOffset(id string, offset int64, progress float64) error
	UpdateStatus(id string, status string) error
	Complete(id string, projectID uint, filePath string) error
	GetExpiredUploads(before time.Time) ([]domain.TusUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusUpload, error)
	Delete(id string) error
	ListActive() ([]domain.TusUpload, error)
	GetActiveUploadIDs() ([]string, error)
}

type TusModulUploadRepository interface {
	Create(upload *domain.TusModulUpload) error
	GetByID(id string) (*domain.TusModulUpload, error)
	GetByUserID(userID string) ([]domain.TusModulUpload, error)
	UpdateOffset(id string, offset int64, progress float64) error
	UpdateStatus(id string, status string) error
	Complete(id string, modulID string, filePath string) error
	Delete(id string) error
	GetExpiredUploads(before time.Time) ([]domain.TusModulUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error)
	CountActiveByUserID(userID string) (int64, error)
	GetActiveByUserID(userID string) ([]domain.TusModulUpload, error)
	GetActiveUploadIDs() ([]string, error)
}
