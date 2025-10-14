package domain_test

import (
	"fiber-boiler-plate/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseResponse_Structure(t *testing.T) {
	response := domain.BaseResponse{
		Success: true,
		Message: "Operasi berhasil",
		Code:    200,
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Operasi berhasil", response.Message)
	assert.Equal(t, 200, response.Code)
}

func TestSuccessResponse_Structure(t *testing.T) {
	now := time.Now()
	data := map[string]interface{}{
		"id":   1,
		"name": "Test",
	}

	response := domain.SuccessResponse{
		BaseResponse: domain.BaseResponse{
			Success: true,
			Message: "Data berhasil diambil",
			Code:    200,
		},
		Data:      data,
		Timestamp: now,
	}

	assert.True(t, response.Success)
	assert.Equal(t, "Data berhasil diambil", response.Message)
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, data, response.Data)
	assert.Equal(t, now, response.Timestamp)
}

func TestErrorResponse_Structure(t *testing.T) {
	now := time.Now()
	errors := []string{"Email wajib diisi", "Password minimal 8 karakter"}

	response := domain.ErrorResponse{
		BaseResponse: domain.BaseResponse{
			Success: false,
			Message: "Validasi gagal",
			Code:    400,
		},
		Errors:    errors,
		Timestamp: now,
	}

	assert.False(t, response.Success)
	assert.Equal(t, "Validasi gagal", response.Message)
	assert.Equal(t, 400, response.Code)
	assert.Equal(t, errors, response.Errors)
	assert.Equal(t, now, response.Timestamp)
}

func TestPaginationData_Structure(t *testing.T) {
	pagination := domain.PaginationData{
		Page:       1,
		Limit:      10,
		TotalItems: 100,
		TotalPages: 10,
	}

	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 10, pagination.Limit)
	assert.Equal(t, 100, pagination.TotalItems)
	assert.Equal(t, 10, pagination.TotalPages)
}

func TestListData_Structure(t *testing.T) {
	items := []map[string]interface{}{
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
	}

	pagination := domain.PaginationData{
		Page:       1,
		Limit:      10,
		TotalItems: 50,
		TotalPages: 5,
	}

	listData := domain.ListData{
		Items:      items,
		Pagination: pagination,
	}

	assert.Equal(t, items, listData.Items)
	assert.Equal(t, pagination, listData.Pagination)
	assert.Equal(t, 1, listData.Pagination.Page)
	assert.Equal(t, 10, listData.Pagination.Limit)
	assert.Equal(t, 50, listData.Pagination.TotalItems)
	assert.Equal(t, 5, listData.Pagination.TotalPages)
}

func TestValidationError_Structure(t *testing.T) {
	validationError := domain.ValidationError{
		Field:   "email",
		Message: "Email tidak valid",
	}

	assert.Equal(t, "email", validationError.Field)
	assert.Equal(t, "Email tidak valid", validationError.Message)
}

func TestValidationError_MultipleErrors(t *testing.T) {
	errors := []domain.ValidationError{
		{Field: "email", Message: "Email wajib diisi"},
		{Field: "password", Message: "Password minimal 8 karakter"},
		{Field: "name", Message: "Nama minimal 2 karakter"},
	}

	assert.Len(t, errors, 3)
	assert.Equal(t, "email", errors[0].Field)
	assert.Equal(t, "Email wajib diisi", errors[0].Message)
	assert.Equal(t, "password", errors[1].Field)
	assert.Equal(t, "Password minimal 8 karakter", errors[1].Message)
	assert.Equal(t, "name", errors[2].Field)
	assert.Equal(t, "Nama minimal 2 karakter", errors[2].Message)
}
