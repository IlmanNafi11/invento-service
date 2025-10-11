package domain

import "time"

type UserListQueryParams struct {
	Search     string `query:"search"`
	FilterRole string `query:"filter_role"`
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
}

type UserListItem struct {
	ID         uint      `json:"id"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	DibuatPada time.Time `json:"dibuat_pada"`
}

type UserListData struct {
	Items      []UserListItem `json:"items"`
	Pagination PaginationData `json:"pagination"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type ProfileData struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UserFileItem struct {
	ID          uint   `json:"id"`
	NamaFile    string `json:"nama_file"`
	Kategori    string `json:"kategori"`
	DownloadURL string `json:"download_url"`
}

type UserFilesQueryParams struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type UserFilesData struct {
	Items      []UserFileItem `json:"items"`
	Pagination PaginationData `json:"pagination"`
}

type UserPermissionItem struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}
