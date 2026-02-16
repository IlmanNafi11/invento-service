package domain

import (
	"testing"
	"time"

	"invento-service/internal/dto"
)

func TestTusModulUploadStruct(t *testing.T) {
	t.Parallel()
	t.Run("TusModulUpload struct initialization", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		modulID := "550e8400-e29b-41d4-a716-446655440050"

		upload := TusModulUpload{
			ID:         "modul-upload-123",
			UserID:     "user-1",
			ModulID:    &modulID,
			UploadType: UploadTypeModulCreate,
			UploadURL:  "https://example.com/upload/modul-upload-123",
			FileSize:   512000,
			FilePath:   "/uploads/module.pdf",
			Status:     UploadStatusPending,
			Progress:   0,
			CreatedAt:  now,
			UpdatedAt:  now,
			ExpiresAt:  now.Add(24 * time.Hour),
		}

		if upload.ID != "modul-upload-123" {
			t.Errorf("Expected ID 'modul-upload-123', got %s", upload.ID)
		}
		if upload.UserID != "user-1" {
			t.Errorf("Expected UserID 'user-1', got %s", upload.UserID)
		}
		if upload.ModulID == nil || *upload.ModulID != "550e8400-e29b-41d4-a716-446655440050" {
			t.Errorf("Expected ModulID '550e8400-e29b-41d4-a716-446655440050', got %v", upload.ModulID)
		}
		if upload.UploadType != UploadTypeModulCreate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeModulCreate, upload.UploadType)
		}
		if upload.Status != UploadStatusPending {
			t.Errorf("Expected Status '%s', got %s", UploadStatusPending, upload.Status)
		}
	})

	t.Run("TusModulUpload without ModulID", func(t *testing.T) {
		t.Parallel()
		upload := TusModulUpload{
			ID:         "modul-upload-456",
			UserID:     "user-2",
			ModulID:    nil,
			UploadType: UploadTypeModulUpdate,
			FileSize:   256000,
			Status:     UploadStatusQueued,
		}

		if upload.ModulID != nil {
			t.Errorf("Expected nil ModulID, got %v", upload.ModulID)
		}
		if upload.UploadType != UploadTypeModulUpdate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeModulUpdate, upload.UploadType)
		}
	})

	t.Run("TusModulUpload with completed status", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		completedAt := now

		upload := TusModulUpload{
			ID:          "modul-upload-789",
			UserID:      "user-3",
			UploadType:  UploadTypeModulCreate,
			FileSize:    1024000,
			Status:      UploadStatusCompleted,
			Progress:    100.0,
			CompletedAt: &completedAt,
		}

		if upload.Status != UploadStatusCompleted {
			t.Errorf("Expected Status '%s', got %s", UploadStatusCompleted, upload.Status)
		}
		if upload.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0, got %f", upload.Progress)
		}
		if upload.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set, got nil")
		}
	})
}

func TestModulUploadStatusConstants(t *testing.T) {
	t.Parallel()
	t.Run("All modul upload status constants are defined", func(t *testing.T) {
		t.Parallel()
		statuses := []string{
			UploadStatusQueued,
			UploadStatusPending,
			UploadStatusUploading,
			UploadStatusCompleted,
			UploadStatusCancelled,
			UploadStatusFailed,
			UploadStatusExpired,
		}

		expectedStatuses := []string{
			"queued",
			"pending",
			"uploading",
			"completed",
			"cancelled",
			"failed",
			"expired",
		}

		for i, status := range statuses {
			if status != expectedStatuses[i] {
				t.Errorf("Expected status '%s', got '%s'", expectedStatuses[i], status)
			}
		}
	})
}

func TestModulUploadTypeConstants(t *testing.T) {
	t.Parallel()
	t.Run("All modul upload type constants are defined", func(t *testing.T) {
		t.Parallel()
		types := []string{
			UploadTypeModulCreate,
			UploadTypeModulUpdate,
		}

		expectedTypes := []string{
			"modul_create",
			"modul_update",
		}

		for i, uploadType := range types {
			if uploadType != expectedTypes[i] {
				t.Errorf("Expected upload type '%s', got '%s'", expectedTypes[i], uploadType)
			}
		}
	})
}

func TestTusModulUploadInitRequest(t *testing.T) {
	t.Parallel()
	t.Run("TusModulUploadMetadata with valid data", func(t *testing.T) {
		t.Parallel()
		req := TusModulUploadMetadata{
			Judul:     "Introduction to Algorithms",
			Deskripsi: "A comprehensive guide to algorithms",
		}

		if req.Judul != "Introduction to Algorithms" {
			t.Errorf("Expected Judul 'Introduction to Algorithms', got %s", req.Judul)
		}
		if req.Deskripsi != "A comprehensive guide to algorithms" {
			t.Errorf("Expected Deskripsi 'A comprehensive guide to algorithms', got %s", req.Deskripsi)
		}
	})

	t.Run("TusModulUploadMetadata with empty deskripsi", func(t *testing.T) {
		t.Parallel()
		req := TusModulUploadMetadata{
			Judul:     "Test Module",
			Deskripsi: "",
		}

		if req.Judul != "Test Module" {
			t.Errorf("Expected Judul 'Test Module', got %s", req.Judul)
		}
		if req.Deskripsi != "" {
			t.Errorf("Expected empty Deskripsi, got %s", req.Deskripsi)
		}
	})
}

func TestTusModulUploadResponse(t *testing.T) {
	t.Parallel()
	t.Run("dto.TusModulUploadResponse struct", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusModulUploadResponse{
			UploadID:  "modul-upload-123",
			UploadURL: "https://example.com/tus/modul-upload-123",
			Offset:    0,
			Length:    512000,
		}

		if resp.UploadID != "modul-upload-123" {
			t.Errorf("Expected UploadID 'modul-upload-123', got %s", resp.UploadID)
		}
		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 512000 {
			t.Errorf("Expected Length 512000, got %d", resp.Length)
		}
	})

	t.Run("dto.TusModulUploadResponse with progress", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusModulUploadResponse{
			UploadID:  "modul-upload-456",
			UploadURL: "https://example.com/tus/modul-upload-456",
			Offset:    256000,
			Length:    512000,
		}

		if resp.Offset != 256000 {
			t.Errorf("Expected Offset 256000, got %d", resp.Offset)
		}
		expectedProgress := float64(resp.Offset) / float64(resp.Length) * 100
		if expectedProgress != 50.0 {
			t.Errorf("Expected progress 50.0, got %f", expectedProgress)
		}
	})
}

func TestTusModulUploadInfoResponse(t *testing.T) {
	t.Parallel()
	t.Run("dto.TusModulUploadInfoResponse struct", func(t *testing.T) {
		t.Parallel()
		now := time.Now()

		resp := dto.TusModulUploadInfoResponse{
			UploadID:  "modul-upload-123",
			ModulID:   "550e8400-e29b-41d4-a716-446655440100",
			Judul:     "Data Structures",
			Deskripsi: "Learn about data structures",
			Status:    UploadStatusUploading,
			Progress:  75.0,
			Offset:    750000,
			Length:    1000000,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if resp.UploadID != "modul-upload-123" {
			t.Errorf("Expected UploadID 'modul-upload-123', got %s", resp.UploadID)
		}
		if resp.ModulID != "550e8400-e29b-41d4-a716-446655440100" {
			t.Errorf("Expected ModulID '550e8400-e29b-41d4-a716-446655440100', got %s", resp.ModulID)
		}
		if resp.Judul != "Data Structures" {
			t.Errorf("Expected Judul 'Data Structures', got %s", resp.Judul)
		}
		if resp.Deskripsi != "Learn about data structures" {
			t.Errorf("Expected Deskripsi 'Learn about data structures', got %s", resp.Deskripsi)
		}
		if resp.Status != UploadStatusUploading {
			t.Errorf("Expected Status '%s', got %s", UploadStatusUploading, resp.Status)
		}
		if resp.Progress != 75.0 {
			t.Errorf("Expected Progress 75.0, got %f", resp.Progress)
		}
	})

	t.Run("dto.TusModulUploadInfoResponse with empty ModulID", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusModulUploadInfoResponse{
			UploadID:  "modul-upload-456",
			ModulID:   "",
			Judul:     "New Module",
			Deskripsi: "A new module description",
			Status:    UploadStatusPending,
			Progress:  0,
		}

		if resp.ModulID != "" {
			t.Errorf("Expected empty ModulID, got %s", resp.ModulID)
		}
		if resp.Judul != "New Module" {
			t.Errorf("Expected Judul 'New Module', got %s", resp.Judul)
		}
	})
}

func TestTusModulUploadSlotResponse(t *testing.T) {
	t.Parallel()
	t.Run("dto.TusModulUploadSlotResponse available", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusModulUploadSlotResponse{
			Available:   true,
			Message:     "Upload slot available",
			QueueLength: 1,
			MaxQueue:    10,
		}

		if !resp.Available {
			t.Error("Expected Available to be true")
		}
		if resp.Message != "Upload slot available" {
			t.Errorf("Expected Message 'Upload slot available', got %s", resp.Message)
		}
		if resp.QueueLength != 1 {
			t.Errorf("Expected QueueLength 1, got %d", resp.QueueLength)
		}
		if resp.MaxQueue != 10 {
			t.Errorf("Expected MaxQueue 10, got %d", resp.MaxQueue)
		}
	})

	t.Run("dto.TusModulUploadSlotResponse not available", func(t *testing.T) {
		t.Parallel()
		resp := dto.TusModulUploadSlotResponse{
			Available:   false,
			Message:     "Queue is full",
			QueueLength: 10,
			MaxQueue:    10,
		}

		if resp.Available {
			t.Error("Expected Available to be false")
		}
		if resp.QueueLength != resp.MaxQueue {
			t.Errorf("Expected QueueLength to equal MaxQueue when full")
		}
	})
}

func TestTusModulUploadProgressCalculation(t *testing.T) {
	t.Parallel()
	t.Run("Calculate progress percentage", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			offset      int64
			length      int64
			expectedPct float64
		}{
			{0, 500000, 0.0},
			{250000, 500000, 50.0},
			{500000, 500000, 100.0},
			{375000, 500000, 75.0},
		}

		for _, tc := range testCases {
			progress := float64(tc.offset) / float64(tc.length) * 100
			if progress != tc.expectedPct {
				t.Errorf("Expected progress %f, got %f", tc.expectedPct, progress)
			}
		}
	})
}

func TestTusModulUploadStatusTransitions(t *testing.T) {
	t.Parallel()
	t.Run("Valid status transitions", func(t *testing.T) {
		t.Parallel()
		transitions := map[string][]string{
			UploadStatusQueued:    {UploadStatusPending},
			UploadStatusPending:   {UploadStatusUploading, UploadStatusCancelled, UploadStatusExpired},
			UploadStatusUploading: {UploadStatusCompleted, UploadStatusFailed, UploadStatusCancelled},
		}

		for from, validTo := range transitions {
			for _, to := range validTo {
				valid := false
				for _, v := range validTo {
					if to == v {
						valid = true
						break
					}
				}
				if !valid {
					t.Errorf("Transition from %s to %s should be valid", from, to)
				}
			}
		}
	})
}

func TestTusModulUploadMetadata(t *testing.T) {
	t.Parallel()
	t.Run("TusModulUpload with metadata", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		metadata := TusModulUploadMetadata{
			Judul:     "Machine Learning Basics",
			Deskripsi: "Introduction to machine learning concepts",
		}

		upload := TusModulUpload{
			ID:             "modul-upload-metadata-1",
			UserID:         "user-1",
			UploadType:     UploadTypeModulCreate,
			UploadMetadata: metadata,
			FileSize:       2500000,
			Status:         UploadStatusPending,
			CreatedAt:      now,
			UpdatedAt:      now,
			ExpiresAt:      now.Add(24 * time.Hour),
		}

		if upload.UploadMetadata.Judul != "Machine Learning Basics" {
			t.Errorf("Expected Judul 'Machine Learning Basics', got %s", upload.UploadMetadata.Judul)
		}
		if upload.UploadMetadata.Deskripsi != "Introduction to machine learning concepts" {
			t.Errorf("Expected Deskripsi 'Introduction to machine learning concepts', got %s", upload.UploadMetadata.Deskripsi)
		}
	})
}
