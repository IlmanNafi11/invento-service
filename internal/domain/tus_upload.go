package domain

import "time"

type TusUpload struct {
	ID            string    `json:"id" gorm:"primaryKey;size:36"`
	UserID        uint      `json:"user_id" gorm:"not null;index"`
	NamaProject   string    `json:"nama_project" gorm:"not null;size:255"`
	Kategori      string    `json:"kategori" gorm:"not null;size:50"`
	Semester      int       `json:"semester" gorm:"not null"`
	FileSize      int64     `json:"file_size" gorm:"not null"`
	CurrentOffset int64     `json:"current_offset" gorm:"default:0"`
	FilePath      string    `json:"file_path" gorm:"not null;size:500"`
	Status        string    `json:"status" gorm:"not null;size:20;index"`
	Progress      float64   `json:"progress" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExpiresAt     time.Time `json:"expires_at" gorm:"index"`
	User          User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

const (
	UploadStatusQueued    = "queued"
	UploadStatusUploading = "uploading"
	UploadStatusCompleted = "completed"
	UploadStatusCancelled = "cancelled"
	UploadStatusFailed    = "failed"
	UploadStatusExpired   = "expired"
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
