package dto

import "time"

// TusUploadInitRequest is the DTO version with validation tags.
// The domain version (TusUploadMetadata) has no validation tags and is used for GORM serialization.
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

// TusModulUploadInitRequest is the DTO version with validation tags.
// The domain version (TusModulUploadMetadata) has no validation tags and is used for GORM serialization.
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
