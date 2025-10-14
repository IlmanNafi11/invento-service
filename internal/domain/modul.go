package domain

import "time"

type Modul struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	NamaFile  string    `json:"nama_file" gorm:"not null;size:255"`
	Tipe      string    `json:"tipe" gorm:"not null;size:50"`
	Ukuran    string    `json:"ukuran" gorm:"not null;size:50"`
	Semester  int       `json:"semester" gorm:"not null;default:1"`
	PathFile  string    `json:"path_file" gorm:"not null;size:500"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ModulCreateRequest struct {
	NamaFile string `form:"nama_file" validate:"required"`
}

type ModulUpdateRequest struct {
	NamaFile string `form:"nama_file" validate:"omitempty"`
}

type ModulListQueryParams struct {
	Search         string `query:"search"`
	FilterType     string `query:"filter_type"`
	FilterSemester int    `query:"filter_semester"`
	Page           int    `query:"page"`
	Limit          int    `query:"limit"`
}

type ModulListItem struct {
	ID                 uint      `json:"id"`
	NamaFile           string    `json:"nama_file"`
	Tipe               string    `json:"tipe"`
	Ukuran             string    `json:"ukuran"`
	Semester           int       `json:"semester"`
	PathFile           string    `json:"path_file"`
	TerakhirDiperbarui time.Time `json:"terakhir_diperbarui"`
}

type ModulListData struct {
	Items      []ModulListItem `json:"items"`
	Pagination PaginationData  `json:"pagination"`
}

type ModulResponse struct {
	ID        uint      `json:"id"`
	NamaFile  string    `json:"nama_file"`
	Tipe      string    `json:"tipe"`
	Ukuran    string    `json:"ukuran"`
	Semester  int       `json:"semester"`
	PathFile  string    `json:"path_file"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ModulCreateResponse struct {
	Items []ModulResponse `json:"items"`
}

type ModulDownloadRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}
