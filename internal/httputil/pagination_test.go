package httputil_test

import (
	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePaginationParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		page          int
		limit         int
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "valid parameters",
			page:          2,
			limit:         20,
			expectedPage:  2,
			expectedLimit: 20,
		},
		{
			name:          "zero page defaults to 1",
			page:          0,
			limit:         10,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "negative page defaults to 1",
			page:          -1,
			limit:         10,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "zero limit defaults to 10",
			page:          1,
			limit:         0,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "negative limit defaults to 10",
			page:          1,
			limit:         -5,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "limit exceeds maximum capped at 100",
			page:          1,
			limit:         200,
			expectedPage:  1,
			expectedLimit: 100,
		},
		{
			name:          "both zero",
			page:          0,
			limit:         0,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "both negative",
			page:          -1,
			limit:         -1,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "limit exactly 100",
			page:          1,
			limit:         100,
			expectedPage:  1,
			expectedLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := httputil.NormalizePaginationParams(tt.page, tt.limit)

			assert.Equal(t, tt.expectedPage, result.Page)
			assert.Equal(t, tt.expectedLimit, result.Limit)
		})
	}
}

func TestCalculatePagination(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		page               int
		limit              int
		totalItems         int
		expectedPage       int
		expectedLimit      int
		expectedTotalPages int
	}{
		{
			name:               "first page with items",
			page:               1,
			limit:              10,
			totalItems:         100,
			expectedPage:       1,
			expectedLimit:      10,
			expectedTotalPages: 10,
		},
		{
			name:               "last page",
			page:               10,
			limit:              10,
			totalItems:         100,
			expectedPage:       10,
			expectedLimit:      10,
			expectedTotalPages: 10,
		},
		{
			name:               "middle page",
			page:               5,
			limit:              10,
			totalItems:         100,
			expectedPage:       5,
			expectedLimit:      10,
			expectedTotalPages: 10,
		},
		{
			name:               "partial last page",
			page:               3,
			limit:              10,
			totalItems:         25,
			expectedPage:       3,
			expectedLimit:      10,
			expectedTotalPages: 3,
		},
		{
			name:               "zero total items",
			page:               1,
			limit:              10,
			totalItems:         0,
			expectedPage:       1,
			expectedLimit:      10,
			expectedTotalPages: 0,
		},
		{
			name:               "fewer items than limit",
			page:               1,
			limit:              10,
			totalItems:         5,
			expectedPage:       1,
			expectedLimit:      10,
			expectedTotalPages: 1,
		},
		{
			name:               "exactly one page",
			page:               1,
			limit:              10,
			totalItems:         10,
			expectedPage:       1,
			expectedLimit:      10,
			expectedTotalPages: 1,
		},
		{
			name:               "page beyond total",
			page:               5,
			limit:              10,
			totalItems:         25,
			expectedPage:       5,
			expectedLimit:      10,
			expectedTotalPages: 3,
		},
		{
			name:               "negative page normalized",
			page:               -1,
			limit:              10,
			totalItems:         100,
			expectedPage:       1,
			expectedLimit:      10,
			expectedTotalPages: 10,
		},
		{
			name:               "large total items",
			page:               1,
			limit:              100,
			totalItems:         1050,
			expectedPage:       1,
			expectedLimit:      100,
			expectedTotalPages: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := httputil.CalculatePagination(tt.page, tt.limit, tt.totalItems)

			assert.Equal(t, tt.expectedPage, result.Page)
			assert.Equal(t, tt.expectedLimit, result.Limit)
			assert.Equal(t, tt.totalItems, result.TotalItems)
			assert.Equal(t, tt.expectedTotalPages, result.TotalPages)
		})
	}
}

func TestCalculatePagination_ReturnsPaginationData(t *testing.T) {
	t.Parallel()
	result := httputil.CalculatePagination(2, 20, 100)

	assert.IsType(t, dto.PaginationData{}, result)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 20, result.Limit)
	assert.Equal(t, 100, result.TotalItems)
	assert.Equal(t, 5, result.TotalPages)
}

func TestCalculateOffset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedOffset int
	}{
		{
			name:           "first page",
			page:           1,
			limit:          10,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			page:           2,
			limit:          10,
			expectedOffset: 10,
		},
		{
			name:           "third page",
			page:           3,
			limit:          20,
			expectedOffset: 40,
		},
		{
			name:           "page zero normalized",
			page:           0,
			limit:          10,
			expectedOffset: 0,
		},
		{
			name:           "negative page normalized",
			page:           -1,
			limit:          10,
			expectedOffset: 0,
		},
		{
			name:           "large page number",
			page:           100,
			limit:          50,
			expectedOffset: 4950,
		},
		{
			name:           "limit zero normalized",
			page:           2,
			limit:          0,
			expectedOffset: 10,
		},
		{
			name:           "limit negative normalized",
			page:           2,
			limit:          -5,
			expectedOffset: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			offset := httputil.CalculateOffset(tt.page, tt.limit)
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

func TestCalculatePagination_Integration(t *testing.T) {
	t.Parallel()
	// Test complete pagination workflow
	totalItems := 95
	limit := 10

	// Calculate pagination for all pages
	totalPages := 10 // 95 items / 10 per page = 9.5, rounded up = 10

	for page := 1; page <= totalPages; page++ {
		t.Run("page_integration", func(t *testing.T) {
			t.Parallel()
			pagination := httputil.CalculatePagination(page, limit, totalItems)
			offset := httputil.CalculateOffset(page, limit)

			// Verify pagination data
			assert.Equal(t, page, pagination.Page)
			assert.Equal(t, limit, pagination.Limit)
			assert.Equal(t, totalItems, pagination.TotalItems)
			assert.Equal(t, totalPages, pagination.TotalPages)

			// Verify offset
			expectedOffset := (page - 1) * limit
			assert.Equal(t, expectedOffset, offset)

			// Verify offset is within bounds
			assert.LessOrEqual(t, offset, totalItems)
		})
	}
}

func TestNormalizePaginationParams_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		page          int
		limit         int
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "very large page",
			page:          999999,
			limit:         10,
			expectedPage:  999999,
			expectedLimit: 10,
		},
		{
			name:          "limit at boundary",
			page:          1,
			limit:         99,
			expectedPage:  1,
			expectedLimit: 99,
		},
		{
			name:          "limit just over boundary",
			page:          1,
			limit:         101,
			expectedPage:  1,
			expectedLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := httputil.NormalizePaginationParams(tt.page, tt.limit)

			assert.Equal(t, tt.expectedPage, result.Page)
			assert.Equal(t, tt.expectedLimit, result.Limit)
		})
	}
}

func TestCalculatePagination_SingleItem(t *testing.T) {
	t.Parallel()
	result := httputil.CalculatePagination(1, 10, 1)

	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 1, result.TotalItems)
	assert.Equal(t, 1, result.TotalPages)
}

func TestCalculatePagination_LargeLimit(t *testing.T) {
	t.Parallel()
	// When limit is capped at 100
	result := httputil.CalculatePagination(1, 200, 500)

	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 100, result.Limit) // Should be capped at 100
	assert.Equal(t, 500, result.TotalItems)
	assert.Equal(t, 5, result.TotalPages) // 500 / 100 = 5
}

func TestCalculateOffset_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedOffset int
	}{
		{
			name:           "page 1 with limit 1",
			page:           1,
			limit:          1,
			expectedOffset: 0,
		},
		{
			name:           "page 2 with limit 1",
			page:           2,
			limit:          1,
			expectedOffset: 1,
		},
		{
			name:           "page 1 with max limit",
			page:           1,
			limit:          100,
			expectedOffset: 0,
		},
		{
			name:           "page 2 with max limit",
			page:           2,
			limit:          100,
			expectedOffset: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			offset := httputil.CalculateOffset(tt.page, tt.limit)
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}
