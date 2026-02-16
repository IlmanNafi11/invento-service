package usecase

import "invento-service/config"

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint) *uint {
	return &u
}

func intPtr(i int) *int {
	return &i
}

// Helper functions for test configs
func getTestModulConfig() *config.Config {
	return &config.Config{
		Upload: config.UploadConfig{
			MaxSize:              524288000, // 500 MB
			MaxSizeModul:         52428800,  // 50 MB
			MaxQueueModulPerUser: 5,
			IdleTimeout:          600, // 10 minutes
		},
	}
}

func getTestTusUploadConfig() *config.Config {
	return &config.Config{
		Upload: config.UploadConfig{
			MaxSize:              524288000, // 500 MB
			MaxSizeProject:       524288000, // 500 MB
			MaxConcurrentProject: 1,
			IdleTimeout:          600, // 10 minutes
		},
	}
}
