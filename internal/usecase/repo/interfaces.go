package repo

import (
	"fiber-boiler-plate/internal/domain"
	"time"
)

type UserRepository interface {
	GetByEmail(email string) (*domain.User, error)
	GetByID(id uint) (*domain.User, error)
	Create(user *domain.User) error
	UpdatePassword(email, hashedPassword string) error
	GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error)
	UpdateRole(userID uint, roleID *uint) error
	UpdateProfile(userID uint, name string, jenisKelamin *string, fotoProfil *string) error
	Delete(userID uint) error
}

type RefreshTokenRepository interface {
	Create(userID uint, token string, expiresAt time.Time) (*domain.RefreshToken, error)
	GetByToken(token string) (*domain.RefreshToken, error)
	RevokeToken(token string) error
	RevokeAllUserTokens(userID uint) error
	CleanupExpired() error
}

type PasswordResetTokenRepository interface {
	Create(email, token string, expiresAt time.Time) (*domain.PasswordResetToken, error)
	GetByToken(token string) (*domain.PasswordResetToken, error)
	MarkAsUsed(token string) error
	CleanupExpired() error
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
	GetByIDs(ids []uint, userID uint) ([]domain.Project, error)
	GetByUserID(userID uint, search string, filterSemester int, filterKategori string, page, limit int) ([]domain.ProjectListItem, int, error)
	CountByUserID(userID uint) (int, error)
	Update(project *domain.Project) error
	Delete(id uint) error
}

type ModulRepository interface {
	Create(modul *domain.Modul) error
	GetByID(id uint) (*domain.Modul, error)
	GetByIDs(ids []uint, userID uint) ([]domain.Modul, error)
	GetByUserID(userID uint, search string, filterType string, filterSemester int, page, limit int) ([]domain.ModulListItem, int, error)
	CountByUserID(userID uint) (int, error)
	Update(modul *domain.Modul) error
	Delete(id uint) error
	UpdateMetadata(modul *domain.Modul) error
}

type TusUploadRepository interface {
	Create(upload *domain.TusUpload) error
	GetByID(id string) (*domain.TusUpload, error)
	GetByUserID(userID uint) ([]domain.TusUpload, error)
	UpdateOffset(id string, offset int64, progress float64) error
	UpdateStatus(id string, status string) error
	GetExpired(before time.Time) ([]domain.TusUpload, error)
	GetByUserIDAndStatus(userID uint, status string) ([]domain.TusUpload, error)
	Delete(id string) error
	ListActive() ([]domain.TusUpload, error)
	UpdateOffsetOnly(id string, offset int64) error
	UpdateUpload(upload *domain.TusUpload) error
}
