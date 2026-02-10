package dto

import (
	"time"
)

// BaseResponse adalah struktur dasar untuk semua response
// Success: menandakan apakah request berhasil
// Message: pesan yang menjelaskan hasil request
// Code: HTTP status code
type BaseResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Berhasil"`
	Code    int    `json:"code" example:"200"`
}

// SuccessResponse adalah response standar untuk request yang berhasil
// Data: payload data yang dikembalikan
// Timestamp: waktu response dibuat
type SuccessResponse struct {
	BaseResponse
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// NewSuccessResponse membuat instance SuccessResponse baru
func NewSuccessResponse(code int, message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
			Code:    code,
		},
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ErrorResponse adalah response standar untuk request yang gagal
// Errors: detail error yang terjadi
// Timestamp: waktu response dibuat
type ErrorResponse struct {
	BaseResponse
	Errors    interface{} `json:"errors,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// NewErrorResponse membuat instance ErrorResponse baru
func NewErrorResponse(code int, message string, errors interface{}) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{
			Success: false,
			Message: message,
			Code:    code,
		},
		Errors:    errors,
		Timestamp: time.Now(),
	}
}

// ValidationErrorResponse adalah response untuk error validasi
// ValidationError: daftar field yang mengalami error validasi
type ValidationErrorResponse struct {
	BaseResponse
	ValidationErrors []ErrorDetail `json:"validation_errors"`
	Timestamp        time.Time     `json:"timestamp"`
}

// NewValidationErrorResponse membuat instance ValidationErrorResponse baru
func NewValidationErrorResponse(message string, errors []ErrorDetail) ValidationErrorResponse {
	return ValidationErrorResponse{
		BaseResponse: BaseResponse{
			Success: false,
			Message: message,
			Code:    400,
		},
		ValidationErrors: errors,
		Timestamp:        time.Now(),
	}
}

// ProblemDetailsResponse adalah response error sesuai RFC 7807 (Problem Details for HTTP APIs)
// Type: URI reference yang mengidentifikasi jenis problem
// Title: judul singkat tentang problem
// Status: HTTP status code
// Detail: penjelasan detail tentang problem
// Instance: URI reference yang mengidentifikasi kejadian spesifik dari problem
type ProblemDetailsResponse struct {
	Type     string      `json:"type,omitempty" example:"https://example.com/probs/out-of-stock"`
	Title    string      `json:"title" example:"Stok Habis"`
	Status   int         `json:"status" example:"400"`
	Detail   string      `json:"detail" example:"Produk yang diminta tidak tersedia"`
	Instance string      `json:"instance,omitempty" example:"/product/123"`
	Errors   interface{} `json:"errors,omitempty"`
}

// NewProblemDetailsResponse membuat instance ProblemDetailsResponse baru
func NewProblemDetailsResponse(status int, title, detail string) ProblemDetailsResponse {
	return ProblemDetailsResponse{
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// ListResponse adalah response standar untuk data berbentuk list/daftar
// Items: daftar item yang dikembalikan
// Pagination: metadata pagination
type ListResponse struct {
	BaseResponse
	Data      interface{}          `json:"data"`
	Pagination PaginationMetadata   `json:"pagination"`
	Timestamp  time.Time            `json:"timestamp"`
}

// NewListResponse membuat instance ListResponse baru
func NewListResponse(code int, message string, items interface{}, pagination PaginationMetadata) ListResponse {
	return ListResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
			Code:    code,
		},
		Data: map[string]interface{}{
			"items": items,
		},
		Pagination: pagination,
		Timestamp:  time.Now(),
	}
}

// MessageResponse adalah response sederhana yang hanya mengandung pesan
type MessageResponse struct {
	BaseResponse
	Timestamp time.Time `json:"timestamp"`
}

// NewMessageResponse membuat instance MessageResponse baru
func NewMessageResponse(code int, message string) MessageResponse {
	return MessageResponse{
		BaseResponse: BaseResponse{
			Success: code >= 200 && code < 300,
			Message: message,
			Code:    code,
		},
		Timestamp: time.Now(),
	}
}

// DataResponse adalah wrapper untuk response dengan struktur data khusus
// ResponseKey: kunci untuk data yang dikembalikan
// ResponseData: data yang dikembalikan
type DataResponse struct {
	ResponseKey   string      `json:"-"`
	ResponseData  interface{} `json:"data,omitempty"`
	BaseResponse
	Timestamp time.Time `json:"timestamp"`
}

// NewDataResponse membuat instance DataResponse baru dengan key khusus
func NewDataResponse(code int, message string, key string, data interface{}) DataResponse {
	return DataResponse{
		ResponseKey: key,
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
			Code:    code,
		},
		ResponseData: data,
		Timestamp:    time.Now(),
	}
}

// MarshalJSON mengimplementasikan custom JSON marshaling untuk DataResponse
func (d DataResponse) MarshalJSON() ([]byte, error) {
	type Alias DataResponse
	if d.ResponseKey != "" && d.ResponseKey != "data" {
		return nil, nil // TODO: implement custom marshaling
	}
	return nil, nil // TODO: implement proper marshaling
}
