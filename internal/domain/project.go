package domain

import "time"

type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	NamaProject string    `json:"nama_project" gorm:"not null;size:255"`
	Kategori    string    `json:"kategori" gorm:"not null;size:50"`
	Semester    int       `json:"semester" gorm:"not null"`
	Ukuran      string    `json:"ukuran" gorm:"not null;size:50"`
	PathFile    string    `json:"path_file" gorm:"not null;size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ProjectCreateRequest struct {
	NamaProject string `form:"nama_project" validate:"required"`
	Semester    int    `form:"semester" validate:"required,min=1,max=8"`
}

type ProjectUpdateRequest struct {
	NamaProject string `form:"nama_project" validate:"omitempty"`
	Semester    int    `form:"semester" validate:"omitempty,min=1,max=8"`
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

type ProjectCreateResponse struct {
	Items []ProjectResponse `json:"items"`
}

type ProjectDownloadRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}
