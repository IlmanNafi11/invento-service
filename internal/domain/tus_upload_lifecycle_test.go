package domain

import (
	"testing"
	"time"
)

func TestTusUpload_Timestamps(t *testing.T) {
	t.Parallel()
	t.Run("Valid timestamps", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		past := now.Add(-1 * time.Hour)
		future := now.Add(24 * time.Hour)

		upload := TusUpload{
			ID:        "upload-timestamps",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			CreatedAt: past,
			UpdatedAt: now,
			ExpiresAt: future,
		}

		if upload.CreatedAt.IsZero() {
			t.Error("Expected non-zero CreatedAt")
		}
		if upload.UpdatedAt.IsZero() {
			t.Error("Expected non-zero UpdatedAt")
		}
		if upload.ExpiresAt.IsZero() {
			t.Error("Expected non-zero ExpiresAt")
		}
	})

	t.Run("Zero CreatedAt", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:        "upload-zero-created",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			CreatedAt: time.Time{},
		}

		if !upload.CreatedAt.IsZero() {
			t.Error("Expected zero CreatedAt")
		}
	})

	t.Run("Zero UpdatedAt", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:        "upload-zero-updated",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			UpdatedAt: time.Time{},
		}

		if !upload.UpdatedAt.IsZero() {
			t.Error("Expected zero UpdatedAt")
		}
	})

	t.Run("Zero ExpiresAt", func(t *testing.T) {
		t.Parallel()
		upload := TusUpload{
			ID:        "upload-zero-expires",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			ExpiresAt: time.Time{},
		}

		if !upload.ExpiresAt.IsZero() {
			t.Error("Expected zero ExpiresAt")
		}
	})

	t.Run("Expired timestamp", func(t *testing.T) {
		t.Parallel()
		past := time.Now().Add(-1 * time.Hour)

		upload := TusUpload{
			ID:        "upload-expired",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusExpired,
			ExpiresAt: past,
		}

		if upload.ExpiresAt.After(time.Now()) {
			t.Error("Expected ExpiresAt to be in the past")
		}
	})

	t.Run("Future expiration", func(t *testing.T) {
		t.Parallel()
		future := time.Now().Add(24 * time.Hour)

		upload := TusUpload{
			ID:        "upload-future-expiration",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			ExpiresAt: future,
		}

		if upload.ExpiresAt.Before(time.Now()) {
			t.Error("Expected ExpiresAt to be in the future")
		}
	})
}
