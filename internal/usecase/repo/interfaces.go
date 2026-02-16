package repo

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"time"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]dto.UserListItem, int, error)
	GetProfileWithCounts(ctx context.Context, userID string) (*domain.User, int, int, error)
	GetUserFiles(ctx context.Context, userID, search string, page, limit int) ([]dto.UserFileItem, int, error)
	UpdateRole(ctx context.Context, userID string, roleID *int) error
	UpdateProfile(ctx context.Context, userID, name string, jenisKelamin, fotoProfil *string) error
	Delete(ctx context.Context, userID string) error
	GetByRoleID(ctx context.Context, roleID uint) ([]dto.UserListItem, error)
	BulkUpdateRole(ctx context.Context, userIDs []string, roleID uint) error
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
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id uint) (*domain.Project, error)
	GetByIDs(ctx context.Context, ids []uint, userID string) ([]domain.Project, error)
	GetByUserID(ctx context.Context, userID, search string, filterSemester int, filterKategori string, page, limit int) ([]dto.ProjectListItem, int, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id uint) error
}

type ModulRepository interface {
	Create(ctx context.Context, modul *domain.Modul) error
	GetByID(ctx context.Context, id string) (*domain.Modul, error)
	GetByIDs(ctx context.Context, ids []string, userID string) ([]domain.Modul, error)
	GetByUserID(ctx context.Context, userID, search, filterType, filterStatus string, page, limit int) ([]dto.ModulListItem, int, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	Update(ctx context.Context, modul *domain.Modul) error
	Delete(ctx context.Context, id string) error
	UpdateMetadata(ctx context.Context, modul *domain.Modul) error
}

type TusUploadRepository interface {
	Create(ctx context.Context, upload *domain.TusUpload) error
	GetByID(ctx context.Context, id string) (*domain.TusUpload, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.TusUpload, error)
	GetActiveByUserID(ctx context.Context, userID string) ([]domain.TusUpload, error)
	CountActiveByUserID(ctx context.Context, userID string) (int64, error)
	UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error
	UpdateStatus(ctx context.Context, id, status string) error
	Complete(ctx context.Context, id string, projectID uint, filePath string) error
	GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusUpload, error)
	GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusUpload, error)
	Delete(ctx context.Context, id string) error
	ListActive(ctx context.Context) ([]domain.TusUpload, error)
	GetActiveUploadIDs(ctx context.Context) ([]string, error)
}

type TusModulUploadRepository interface {
	Create(ctx context.Context, upload *domain.TusModulUpload) error
	GetByID(ctx context.Context, id string) (*domain.TusModulUpload, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.TusModulUpload, error)
	UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error
	UpdateStatus(ctx context.Context, id, status string) error
	Complete(ctx context.Context, id, modulID, filePath string) error
	Delete(ctx context.Context, id string) error
	GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusModulUpload, error)
	GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusModulUpload, error)
	CountActiveByUserID(ctx context.Context, userID string) (int64, error)
	GetActiveByUserID(ctx context.Context, userID string) ([]domain.TusModulUpload, error)
	GetActiveUploadIDs(ctx context.Context) ([]string, error)
}
