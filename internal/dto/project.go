package dto

import "time"

// CreateProjectRequest follows Action+Entity+Type naming convention.
// Renamed from domain.ProjectCreateRequest.
type CreateProjectRequest struct {
	NamaProject string `form:"nama_project" validate:"required"`
	Semester    int    `form:"semester" validate:"required,min=1,max=8"`
}

// UpdateProjectRequest follows Action+Entity+Type naming convention.
// Renamed from domain.ProjectUpdateRequest.
type UpdateProjectRequest struct {
	NamaProject string `json:"nama_project" validate:"omitempty,min=3,max=255"`
	Kategori    string `json:"kategori" validate:"omitempty,oneof=website mobile iot machine_learning deep_learning"`
	Semester    int    `json:"semester" validate:"omitempty,min=1,max=8"`
}

type ProjectListQueryParams struct {
	Search         string `query:"search"`
	FilterSemester int    `query:"filter_semester"`
	FilterKategori string `query:"filter_kategori"`
	Page           int    `query:"page"`
	Limit          int    `query:"limit"`
}

type ProjectListItem struct {
	ID                 uint      `json:"id"`
	NamaProject        string    `json:"nama_project"`
	Kategori           string    `json:"kategori"`
	Semester           int       `json:"semester"`
	Ukuran             string    `json:"ukuran"`
	PathFile           string    `json:"path_file"`
	TerakhirDiperbarui time.Time `json:"terakhir_diperbarui"`
}

type ProjectListData struct {
	Items      []ProjectListItem `json:"items"`
	Pagination PaginationData    `json:"pagination"`
}

type ProjectResponse struct {
	ID          uint      `json:"id"`
	NamaProject string    `json:"nama_project"`
	Kategori    string    `json:"kategori"`
	Semester    int       `json:"semester"`
	Ukuran      string    `json:"ukuran"`
	PathFile    string    `json:"path_file"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectDownloadRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}
