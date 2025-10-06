package domain

import "time"

type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	NamaRole  string    `json:"nama_role" gorm:"uniqueIndex;not null;size:50"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Resource  string    `json:"resource" gorm:"not null;size:100"`
	Action    string    `json:"action" gorm:"not null;size:50"`
	Label     string    `json:"label" gorm:"not null;size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RolePermission struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	RoleID       uint       `json:"role_id" gorm:"not null"`
	PermissionID uint       `json:"permission_id" gorm:"not null"`
	CreatedAt    time.Time  `json:"created_at"`
	Role         Role       `json:"role" gorm:"foreignKey:RoleID"`
	Permission   Permission `json:"permission" gorm:"foreignKey:PermissionID"`
}

type RoleCreateRequest struct {
	NamaRole    string                       `json:"nama_role" validate:"required,min=2,max=50"`
	Permissions map[string][]string          `json:"permissions" validate:"required"`
}

type RoleUpdateRequest struct {
	NamaRole    string                       `json:"nama_role" validate:"required,min=2,max=50"`
	Permissions map[string][]string          `json:"permissions" validate:"required"`
}

type RoleListQueryParams struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type RoleListItem struct {
	ID                 uint      `json:"id"`
	NamaRole           string    `json:"nama_role"`
	JumlahPermission   int       `json:"jumlah_permission"`
	TanggalDiperbarui  time.Time `json:"tanggal_diperbarui"`
}

type RolePermissionDetail struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

type RoleDetailResponse struct {
	ID               uint                   `json:"id"`
	NamaRole         string                 `json:"nama_role"`
	Permissions      []RolePermissionDetail `json:"permissions"`
	JumlahPermission int                    `json:"jumlah_permission"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type PermissionItem struct {
	Action string `json:"action"`
	Label  string `json:"label"`
}

type ResourcePermissions struct {
	Name        string           `json:"name"`
	Permissions []PermissionItem `json:"permissions"`
}

type PaginationData struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type RoleListData struct {
	Items      []RoleListItem `json:"items"`
	Pagination PaginationData `json:"pagination"`
}
