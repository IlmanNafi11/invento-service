package helper_test

import (
	"fiber-boiler-plate/internal/helper"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockRefreshTokenRepository is defined in auth_helper_test.go to avoid duplication

func TestNewRefreshTokenCleanup(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	intervalHours := 24

	cleanup := helper.NewRefreshTokenCleanup(mockRepo, intervalHours)

	assert.NotNil(t, cleanup)
}

func TestRefreshTokenCleanup_Start(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	mockRepo.On("CleanupExpired").Return(nil).Maybe()
	cleanup := helper.NewRefreshTokenCleanup(mockRepo, 1)

	cleanup.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the cleanup to avoid goroutine leak
	cleanup.Stop()
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

func TestRefreshTokenCleanup_Stop(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	cleanup := helper.NewRefreshTokenCleanup(mockRepo, 1)

	cleanup.Start()
	time.Sleep(100 * time.Millisecond)

	// This should not panic or block indefinitely
	cleanup.Stop()
	time.Sleep(100 * time.Millisecond)
}

func TestRefreshTokenCleanup_StartStopMultiple(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	mockRepo.On("CleanupExpired").Return(nil).Maybe()
	cleanup := helper.NewRefreshTokenCleanup(mockRepo, 1)

	// Test multiple start/stop cycles
	for i := 0; i < 3; i++ {
		cleanup.Start()
		time.Sleep(50 * time.Millisecond)
		cleanup.Stop()
		time.Sleep(50 * time.Millisecond)
	}

	mockRepo.AssertExpectations(t)
}

func TestNewRefreshTokenCleanup_DefaultValues(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)

	tests := []struct {
		name          string
		intervalHours int
	}{
		{"24 hours", 24},
		{"1 hour", 1},
		{"12 hours", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := helper.NewRefreshTokenCleanup(mockRepo, tt.intervalHours)
			assert.NotNil(t, cleanup)
		})
	}
}

func TestRefreshTokenCleanup_Cleanup(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	mockRepo.On("CleanupExpired").Return(nil).Maybe()

	cleanup := helper.NewRefreshTokenCleanup(mockRepo, 1)

	// Start cleanup
	cleanup.Start()
	time.Sleep(100 * time.Millisecond)

	// Stop and wait for cleanup to potentially run
	cleanup.Stop()
	time.Sleep(200 * time.Millisecond)

	// The cleanup may or may not have run depending on timing
	mockRepo.AssertExpectations(t)
}
