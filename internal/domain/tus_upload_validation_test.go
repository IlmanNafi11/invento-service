package domain

import (
	"testing"

	"invento-service/internal/dto"
)

func TestTusUploadResponse_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Zero offset and length", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadResponse{
			UploadID:  "upload-zero",
			UploadURL: "https://example.com/tus/upload-zero",
			Offset:    0,
			Length:    0,
		}

		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 0 {
			t.Errorf("Expected Length 0, got %d", resp.Length)
		}
	})

	t.Run("Empty upload ID", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadResponse{
			UploadID:  "",
			UploadURL: "https://example.com/tus/",
			Offset:    0,
			Length:    1000,
		}

		if resp.UploadID != "" {
			t.Errorf("Expected empty UploadID, got '%s'", resp.UploadID)
		}
	})

	t.Run("Empty upload URL", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadResponse{
			UploadID:  "upload-123",
			UploadURL: "",
			Offset:    0,
			Length:    1000,
		}

		if resp.UploadURL != "" {
			t.Errorf("Expected empty UploadURL, got '%s'", resp.UploadURL)
		}
	})

	t.Run("Large file size", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadResponse{
			UploadID:  "upload-large",
			UploadURL: "https://example.com/tus/upload-large",
			Offset:    0,
			Length:    1024 * 1024 * 1024 * 5, // 5GB
		}

		expectedLength := int64(1024 * 1024 * 1024 * 5)
		if resp.Length != expectedLength {
			t.Errorf("Expected Length %d, got %d", expectedLength, resp.Length)
		}
	})
}

func TestTusUploadInfoResponse_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Zero ProjectID", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-zero-project",
			ProjectID:   0,
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    1,
			Status:      UploadStatusPending,
			Progress:    0,
		}

		if resp.ProjectID != 0 {
			t.Errorf("Expected ProjectID 0, got %d", resp.ProjectID)
		}
	})

	t.Run("Invalid progress - negative value", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-negative-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "mobile",
			Semester:    2,
			Status:      UploadStatusUploading,
			Progress:    -10.5,
		}

		if resp.Progress != -10.5 {
			t.Errorf("Expected Progress -10.5, got %f", resp.Progress)
		}
	})

	t.Run("Invalid progress - exceeds 100", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-excess-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "iot",
			Semester:    3,
			Status:      UploadStatusUploading,
			Progress:    150.0,
		}

		if resp.Progress != 150.0 {
			t.Errorf("Expected Progress 150.0, got %f", resp.Progress)
		}
	})

	t.Run("Zero progress", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-zero-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "machine_learning",
			Semester:    4,
			Status:      UploadStatusPending,
			Progress:    0,
		}

		if resp.Progress != 0 {
			t.Errorf("Expected Progress 0, got %f", resp.Progress)
		}
	})

	t.Run("Exact 100 percent progress", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-complete-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "deep_learning",
			Semester:    5,
			Status:      UploadStatusCompleted,
			Progress:    100.0,
		}

		if resp.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0, got %f", resp.Progress)
		}
	})

	t.Run("Zero offset and length", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-zero-offset",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    1,
			Status:      UploadStatusPending,
			Progress:    0,
			Offset:      0,
			Length:      0,
		}

		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 0 {
			t.Errorf("Expected Length 0, got %d", resp.Length)
		}
	})
}

func TestTusUploadSlotResponse_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Zero queue length", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadSlotResponse{
			Available:     true,
			Message:       "No uploads in queue",
			QueueLength:   0,
			ActiveUpload:  false,
			MaxConcurrent: 5,
		}

		if resp.QueueLength != 0 {
			t.Errorf("Expected QueueLength 0, got %d", resp.QueueLength)
		}
	})

	t.Run("Zero max concurrent", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadSlotResponse{
			Available:     false,
			Message:       "Uploads disabled",
			QueueLength:   0,
			ActiveUpload:  false,
			MaxConcurrent: 0,
		}

		if resp.MaxConcurrent != 0 {
			t.Errorf("Expected MaxConcurrent 0, got %d", resp.MaxConcurrent)
		}
	})

	t.Run("Empty message", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadSlotResponse{
			Available:     true,
			Message:       "",
			QueueLength:   1,
			ActiveUpload:  true,
			MaxConcurrent: 5,
		}

		if resp.Message != "" {
			t.Errorf("Expected empty Message, got '%s'", resp.Message)
		}
	})

	t.Run("Large queue length", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusUploadSlotResponse{
			Available:     false,
			Message:       "Very long queue",
			QueueLength:   1000,
			ActiveUpload:  true,
			MaxConcurrent: 5,
		}

		if resp.QueueLength != 1000 {
			t.Errorf("Expected QueueLength 1000, got %d", resp.QueueLength)
		}
	})
}

func TestTusUpload_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Nil ProjectID", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-no-project",
			UserID:     "user-1",
			ProjectID:  nil,
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.ProjectID != nil {
			t.Errorf("Expected nil ProjectID, got %v", upload.ProjectID)
		}
	})

	t.Run("Zero file size", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-zero-size",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   0,
			Status:     UploadStatusPending,
		}

		if upload.FileSize != 0 {
			t.Errorf("Expected FileSize 0, got %d", upload.FileSize)
		}
	})

	t.Run("Negative file size", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-negative-size",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   -1,
			Status:     UploadStatusPending,
		}

		if upload.FileSize != -1 {
			t.Errorf("Expected FileSize -1, got %d", upload.FileSize)
		}
	})

	t.Run("Empty upload ID", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.ID != "" {
			t.Errorf("Expected empty ID, got '%s'", upload.ID)
		}
	})

	t.Run("Zero UserID", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-zero-user",
			UserID:     "",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.UserID != "" {
			t.Errorf("Expected UserID empty, got %s", upload.UserID)
		}
	})

	t.Run("Empty upload URL", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-empty-url",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			UploadURL:  "",
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.UploadURL != "" {
			t.Errorf("Expected empty UploadURL, got '%s'", upload.UploadURL)
		}
	})

	t.Run("Empty file path", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-empty-path",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			FilePath:   "",
			Status:     UploadStatusPending,
		}

		if upload.FilePath != "" {
			t.Errorf("Expected empty FilePath, got '%s'", upload.FilePath)
		}
	})

	t.Run("Zero current offset", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:            "upload-zero-offset",
			UserID:        "user-1",
			UploadType:    UploadTypeProjectCreate,
			FileSize:      1024,
			CurrentOffset: 0,
			Status:        UploadStatusPending,
		}

		if upload.CurrentOffset != 0 {
			t.Errorf("Expected CurrentOffset 0, got %d", upload.CurrentOffset)
		}
	})

	t.Run("Negative current offset", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:            "upload-negative-offset",
			UserID:        "user-1",
			UploadType:    UploadTypeProjectCreate,
			FileSize:      1024,
			CurrentOffset: -1,
			Status:        UploadStatusPending,
		}

		if upload.CurrentOffset != -1 {
			t.Errorf("Expected CurrentOffset -1, got %d", upload.CurrentOffset)
		}
	})

	t.Run("Zero progress", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-zero-progress",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
			Progress:   0,
		}

		if upload.Progress != 0 {
			t.Errorf("Expected Progress 0, got %f", upload.Progress)
		}
	})

	t.Run("Negative progress", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-negative-progress",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusUploading,
			Progress:   -5.5,
		}

		if upload.Progress != -5.5 {
			t.Errorf("Expected Progress -5.5, got %f", upload.Progress)
		}
	})

	t.Run("Progress exceeds 100", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-excess-progress",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusUploading,
			Progress:   120.5,
		}

		if upload.Progress != 120.5 {
			t.Errorf("Expected Progress 120.5, got %f", upload.Progress)
		}
	})

	t.Run("Nil CompletedAt", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:          "upload-no-completed",
			UserID:      "user-1",
			UploadType:  UploadTypeProjectCreate,
			FileSize:    1024,
			Status:      UploadStatusPending,
			CompletedAt: nil,
		}

		if upload.CompletedAt != nil {
			t.Errorf("Expected nil CompletedAt, got %v", upload.CompletedAt)
		}
	})

	t.Run("Empty upload type", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-empty-type",
			UserID:     "user-1",
			UploadType: "",
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.UploadType != "" {
			t.Errorf("Expected empty UploadType, got '%s'", upload.UploadType)
		}
	})

	t.Run("Empty status", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:         "upload-empty-status",
			UserID:     "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     "",
		}

		if upload.Status != "" {
			t.Errorf("Expected empty Status, got '%s'", upload.Status)
		}
	})
}
