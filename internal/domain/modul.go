package domain

import (
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

func (m *Modul) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

func (Modul) TableName() string {
	return "moduls"
}
