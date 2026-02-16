package domain

import (
	"invento-service/internal/dto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Modul struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;type:uuid"`
	Judul     string    `json:"judul" gorm:"not null;size:255"`
	Deskripsi string    `json:"deskripsi" gorm:"type:text"`
	FilePath  string    `json:"file_path" gorm:"column:file_path;size:500"`
	FileName  string    `json:"file_name" gorm:"column:file_name;size:255"`
	FileSize  int64     `json:"file_size" gorm:"column:file_size"`
	MimeType  string    `json:"mime_type" gorm:"column:mime_type;size:100"`
	Status    string    `json:"status" gorm:"size:50;default:'pending'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate generates a UUID if not provided
func (m *Modul) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

func (Modul) TableName() string {
	return "moduls"
}

type ModulUpdateRequest struct {
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
	Items      []ModulListItem    `json:"items"`
	Pagination dto.PaginationData `json:"pagination"`
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
