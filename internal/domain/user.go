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
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	JenisKelamin  *string   `json:"jenis_kelamin,omitempty"`
	FotoProfil    *string   `json:"foto_profil,omitempty"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	JumlahProject int       `json:"jumlah_project"`
	JumlahModul   int       `json:"jumlah_modul"`
}

type UpdateProfileRequest struct {
	Name         string `form:"name" validate:"required,min=2,max=100"`
	JenisKelamin string `form:"jenis_kelamin" validate:"omitempty,oneof=Laki-laki Perempuan"`
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

type DownloadUserFilesRequest struct {
	ProjectIDs []uint `json:"project_ids" validate:"omitempty"`
	ModulIDs   []uint `json:"modul_ids" validate:"omitempty"`
}
