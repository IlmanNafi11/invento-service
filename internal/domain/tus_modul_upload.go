package domain

import "time"

type TusModulUpload struct {
	ID             string                    `json:"id" gorm:"primaryKey;size:36"`
	UserID         string                    `json:"user_id" gorm:"not null;index;type:uuid"`
	ModulID        *uint                     `json:"modul_id,omitempty" gorm:"index"`
	UploadType     string                    `json:"upload_type" gorm:"not null;size:20;default:'modul_create'"`
	UploadURL      string                    `json:"upload_url" gorm:"size:500"`
	UploadMetadata TusModulUploadInitRequest `json:"upload_metadata" gorm:"serializer:json"`
	FileSize       int64                     `json:"file_size" gorm:"not null"`
	CurrentOffset  int64                     `json:"current_offset" gorm:"default:0"`
	FilePath       string                    `json:"file_path" gorm:"size:500"`
	Status         string                    `json:"status" gorm:"not null;size:20;index"`
	Progress       float64                   `json:"progress" gorm:"default:0"`
	CompletedAt    *time.Time                `json:"completed_at,omitempty"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
	ExpiresAt      time.Time                 `json:"expires_at" gorm:"index"`
	User           User                      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

const (
	ModulUploadStatusQueued    = "queued"
	ModulUploadStatusPending   = "pending"
	ModulUploadStatusUploading = "uploading"
	ModulUploadStatusCompleted = "completed"
	ModulUploadStatusCancelled = "cancelled"
	ModulUploadStatusFailed    = "failed"
	ModulUploadStatusExpired   = "expired"

	ModulUploadTypeCreate = "modul_create"
	ModulUploadTypeUpdate = "modul_update"
)

type TusModulUploadInitRequest struct {
	NamaFile string `json:"nama_file" validate:"required,min=3,max=255"`
	Tipe     string `json:"tipe" validate:"required,oneof=docx xlsx pdf pptx"`
	Semester int    `json:"semester" validate:"required,min=1,max=8"`
}

type TusModulUploadResponse struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
	Offset    int64  `json:"offset"`
	Length    int64  `json:"length"`
}

type TusModulUploadInfoResponse struct {
	UploadID  string    `json:"upload_id"`
	ModulID   uint      `json:"modul_id,omitempty"`
	NamaFile  string    `json:"nama_file"`
	Tipe      string    `json:"tipe"`
	Semester  int       `json:"semester"`
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

type ModulUpdateMetadataRequest struct {
	NamaFile string `json:"nama_file" validate:"required,min=3,max=255"`
	Semester int    `json:"semester" validate:"required,min=1,max=8"`
}
