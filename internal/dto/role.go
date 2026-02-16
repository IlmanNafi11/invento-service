package dto

import "time"

type RoleCreateRequest struct {
	NamaRole    string              `json:"nama_role" validate:"required,min=2,max=50"`
	Permissions map[string][]string `json:"permissions" validate:"required"`
}

type RoleUpdateRequest struct {
	NamaRole    string              `json:"nama_role" validate:"required,min=2,max=50"`
	Permissions map[string][]string `json:"permissions" validate:"required"`
}

type RoleListQueryParams struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type RoleListItem struct {
	ID                uint      `json:"id"`
	NamaRole          string    `json:"nama_role"`
	JumlahPermission  int       `json:"jumlah_permission"`
	TanggalDiperbarui time.Time `json:"tanggal_diperbarui"`
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

type RoleListData struct {
	Items      []RoleListItem `json:"items"`
	Pagination PaginationData `json:"pagination"`
}
