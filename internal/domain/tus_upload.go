package domain

import "time"

type TusUpload struct {
	ID             string               `json:"id" gorm:"primaryKey;size:36"`
	UserID         string               `json:"user_id" gorm:"not null;index;type:uuid"`
	ProjectID      *uint                `json:"project_id,omitempty" gorm:"index"`
	UploadType     string               `json:"upload_type" gorm:"not null;size:20;default:'project_create'"`
	UploadURL      string               `json:"upload_url" gorm:"size:500"`
	UploadMetadata TusUploadInitRequest `json:"upload_metadata" gorm:"serializer:json"`
	FileSize       int64                `json:"file_size" gorm:"not null"`
	CurrentOffset  int64                `json:"current_offset" gorm:"default:0"`
	FilePath       string               `json:"file_path" gorm:"size:500"`
	Status         string               `json:"status" gorm:"not null;size:20;index"`
	Progress       float64              `json:"progress" gorm:"default:0"`
	CompletedAt    *time.Time           `json:"completed_at,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	ExpiresAt      time.Time            `json:"expires_at" gorm:"index"`
	User           User                 `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

const (
	UploadStatusQueued    = "queued"
	UploadStatusPending   = "pending"
	UploadStatusUploading = "uploading"
	UploadStatusCompleted = "completed"
	UploadStatusCancelled = "cancelled"
	UploadStatusFailed    = "failed"
	UploadStatusExpired   = "expired"

	UploadTypeProjectCreate = "project_create"
	UploadTypeProjectUpdate = "project_update"
)

type TusUploadInitRequest struct {
	NamaProject string `json:"nama_project" validate:"required,min=3,max=255"`
	Kategori    string `json:"kategori" validate:"required,oneof=website mobile iot machine_learning deep_learning"`
	Semester    int    `json:"semester" validate:"required,min=1,max=8"`
}

type TusUploadResponse struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
	Offset    int64  `json:"offset"`
	Length    int64  `json:"length"`
}

type TusUploadInfoResponse struct {
	UploadID    string    `json:"upload_id"`
	ProjectID   uint      `json:"project_id,omitempty"`
	NamaProject string    `json:"nama_project"`
	Kategori    string    `json:"kategori"`
	Semester    int       `json:"semester"`
	Status      string    `json:"status"`
	Progress    float64   `json:"progress"`
	Offset      int64     `json:"offset"`
	Length      int64     `json:"length"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TusUploadSlotResponse struct {
	Available     bool   `json:"available"`
	Message       string `json:"message"`
	QueueLength   int    `json:"queue_length"`
	ActiveUpload  bool   `json:"active_upload"`
	MaxConcurrent int    `json:"max_concurrent"`
}
