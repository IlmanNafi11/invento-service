package domain

import "time"

type BaseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type SuccessResponse struct {
	BaseResponse
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorResponse struct {
	BaseResponse
	Errors    interface{} `json:"errors,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type PaginationData struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type ListData struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationData `json:"pagination"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
