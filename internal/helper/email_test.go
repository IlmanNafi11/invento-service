package helper_test

import (
	"invento-service/internal/helper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePolijeEmail_Valid(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		expectedRole string
		expectedSub  string
	}{
		{
			name:         "student email",
			email:        "test@student.polije.ac.id",
			expectedRole: "mahasiswa",
			expectedSub:  "student",
		},
		{
			name:         "teacher email",
			email:        "test@teacher.polije.ac.id",
			expectedRole: "dosen",
			expectedSub:  "teacher",
		},
		{
			name:         "student with uppercase",
			email:        "TEST@STUDENT.POLIJE.AC.ID",
			expectedRole: "mahasiswa",
			expectedSub:  "student",
		},
		{
			name:         "student with spaces",
			email:        "  test@student.polije.ac.id  ",
			expectedRole: "mahasiswa",
			expectedSub:  "student",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := helper.ValidatePolijeEmail(tt.email)

			assert.NoError(t, err)
			assert.NotNil(t, info)
			assert.True(t, info.IsValid)
			assert.Equal(t, tt.expectedRole, info.RoleName)
			assert.Equal(t, tt.expectedSub, info.Subdomain)
		})
	}
}

func TestValidatePolijeEmail_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		errorMsg string
	}{
		{
			name:     "missing @",
			email:    "testpolije.ac.id",
			errorMsg: "format email tidak valid",
		},
		{
			name:     "multiple @",
			email:    "test@@polije.ac.id",
			errorMsg: "format email tidak valid",
		},
		{
			name:     "wrong domain",
			email:    "test@gmail.com",
			errorMsg: "hanya email dengan domain polije.ac.id",
		},
		{
			name:     "wrong subdomain",
			email:    "test@admin.polije.ac.id",
			errorMsg: "subdomain email tidak valid",
		},
		{
			name:     "no subdomain",
			email:    "test@polije.ac.id",
			errorMsg: "subdomain email tidak valid",
		},
		{
			name:     "empty email",
			email:    "",
			errorMsg: "format email tidak valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := helper.ValidatePolijeEmail(tt.email)

			assert.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestValidatePolijeEmail_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		isValid  bool
	}{
		{
			name:    "student with dots in username",
			email:   "first.last@student.polije.ac.id",
			isValid: true,
		},
		{
			name:    "student with numbers",
			email:   "user123@student.polije.ac.id",
			isValid: true,
		},
		{
			name:    "teacher with dots",
			email:   "john.doe@teacher.polije.ac.id",
			isValid: true,
		},
		{
			name:    "similar but wrong domain",
			email:   "test@polije.ac.id.com",
			isValid: false,
		},
		{
			name:    "partial domain match",
			email:   "test@notpolije.ac.id",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := helper.ValidatePolijeEmail(tt.email)

			if tt.isValid {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.True(t, info.IsValid)
			} else {
				assert.Error(t, err)
				assert.Nil(t, info)
			}
		})
	}
}

func TestValidatePolijeEmail_SubdomainExtraction(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		expectedSub  string
		expectError  bool
	}{
		{
			name:        "student subdomain",
			email:       "user@student.polije.ac.id",
			expectedSub: "student",
			expectError: false,
		},
		{
			name:        "teacher subdomain",
			email:       "user@teacher.polije.ac.id",
			expectedSub: "teacher",
			expectError: false,
		},
		{
			name:        "complex subdomain - invalid",
			email:       "user@sub.student.polije.ac.id",
			expectedSub: "sub",
			expectError: true, // "sub" is not a valid subdomain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := helper.ValidatePolijeEmail(tt.email)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSub, info.Subdomain)
			}
		})
	}
}
