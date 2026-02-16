package dto

import "time"

// UpdateModulRequest follows Action+Entity+Type naming convention.
// Renamed from domain.ModulUpdateRequest.
type UpdateModulRequest struct {
	Judul     string `json:"judul" validate:"omitempty,min=3,max=255"`
	Deskripsi string `json:"deskripsi"`
}

type ModulListQueryParams struct {
	Search       string `query:"search"`
	FilterType   string `query:"filter_type"`
	FilterStatus string `query:"filter_status"`
	Page         int    `query:"page"`
	Limit        int    `query:"limit"`
}

type ModulListItem struct {
	ID                 string    `json:"id"`
	Judul              string    `json:"judul"`
	Deskripsi          string    `json:"deskripsi"`
	FileName           string    `json:"file_name"`
	MimeType           string    `json:"mime_type"`
	FileSize           int64     `json:"file_size"`
	Status             string    `json:"status"`
	TerakhirDiperbarui time.Time `json:"terakhir_diperbarui"`
}

type ModulListData struct {
	Items      []ModulListItem `json:"items"`
	Pagination PaginationData  `json:"pagination"`
}

type ModulResponse struct {
	ID        string    `json:"id"`
	Judul     string    `json:"judul"`
	Deskripsi string    `json:"deskripsi"`
	FileName  string    `json:"file_name"`
	MimeType  string    `json:"mime_type"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ModulDownloadRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}
