package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPaginationRequest_GetOffset_Success tests offset calculation
func TestPaginationRequest_GetOffset_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{"Page 1, Limit 10", 1, 10, 0},
		{"Page 2, Limit 10", 2, 10, 10},
		{"Page 3, Limit 20", 3, 20, 40},
		{"Page 5, Limit 5", 5, 5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := PaginationRequest{
				Page:  tt.page,
				Limit: tt.limit,
			}
			assert.Equal(t, tt.expected, req.GetOffset())
		})
	}
}

// TestPaginationRequest_Validate_Success tests valid pagination request
func TestPaginationRequest_Validate_Success(t *testing.T) {
	t.Parallel()
	req := PaginationRequest{
		Page:   1,
		Limit:  10,
		Sort:   "id",
		Order:  "asc",
	}

	err := req.Validate()
	assert.NoError(t, err)
}

// TestPaginationRequest_Validate_Failure tests invalid pagination request
func TestPaginationRequest_Validate_Failure(t *testing.T) {
	t.Parallel()
	req := PaginationRequest{
		Page:   0, // Invalid: min=1
		Limit:  10,
		Sort:   "id",
		Order:  "asc",
	}

	err := req.Validate()
	assert.Error(t, err)
}

// TestPaginationRequest_Validate_InvalidOrder tests invalid order parameter
func TestPaginationRequest_Validate_InvalidOrder(t *testing.T) {
	t.Parallel()
	req := PaginationRequest{
		Page:   1,
		Limit:  10,
		Sort:   "id",
		Order:  "invalid", // Should be "asc" or "desc"
	}

	err := req.Validate()
	assert.Error(t, err)
}

// TestPaginationRequest_Validate_LimitTooLarge tests limit exceeding max
func TestPaginationRequest_Validate_LimitTooLarge(t *testing.T) {
	t.Parallel()
	req := PaginationRequest{
		Page:   1,
		Limit:  101, // Invalid: max=100
		Sort:   "id",
		Order:  "asc",
	}

	err := req.Validate()
	assert.Error(t, err)
}

// TestIDParam_Validate_Success tests valid ID param
func TestIDParam_Validate_Success(t *testing.T) {
	t.Parallel()
	param := IDParam{
		ID: 123,
	}

	err := param.Validate()
	assert.NoError(t, err)
}

// TestIDParam_Validate_ZeroID tests zero ID which should fail required validation
func TestIDParam_Validate_ZeroID(t *testing.T) {
	t.Parallel()
	param := IDParam{
		ID: 0, // Zero value, should fail required validation
	}

	err := param.Validate()
	assert.Error(t, err)
}
