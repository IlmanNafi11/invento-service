package domain

import "time"

type TusModulUpload struct {
	ID             string                 `json:"id" gorm:"primaryKey;size:36"`
	UserID         string                 `json:"user_id" gorm:"not null;index;type:uuid"`
	ModulID        *string                `json:"modul_id,omitempty" gorm:"index;type:uuid"`
	UploadType     string                 `json:"upload_type" gorm:"not null;size:20;default:'modul_create'"`
	UploadURL      string                 `json:"upload_url" gorm:"size:500"`
	UploadMetadata TusModulUploadMetadata `json:"upload_metadata" gorm:"serializer:json"`
	FileSize       int64                  `json:"file_size" gorm:"not null"`
	CurrentOffset  int64                  `json:"current_offset" gorm:"default:0"`
	FilePath       string                 `json:"file_path" gorm:"size:500"`
	Status         string                 `json:"status" gorm:"not null;size:20;index"`
	Progress       float64                `json:"progress" gorm:"default:0"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ExpiresAt      time.Time              `json:"expires_at" gorm:"index"`
	User           User                   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type TusModulUploadMetadata struct {
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
}

type TusModulUploadInitRequest struct {
	Judul     string `json:"judul" validate:"required,min=3,max=255"`
	Deskripsi string `json:"deskripsi"`
}

type TusModulUploadResponse struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
	Offset    int64  `json:"offset"`
	Length    int64  `json:"length"`
}

type TusModulUploadInfoResponse struct {
	UploadID  string    `json:"upload_id"`
	ModulID   string    `json:"modul_id,omitempty"`
	Judul     string    `json:"judul"`
	Deskripsi string    `json:"deskripsi"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	Offset    int64     `json:"offset"`
	Length    int64     `json:"length"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TusModulUploadSlotResponse struct {
	Available   bool   `json:"available"`
	Message     string `json:"message"`
	QueueLength int    `json:"queue_length"`
	MaxQueue    int    `json:"max_queue"`
}
