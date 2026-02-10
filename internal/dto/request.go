package dto

import "github.com/go-playground/validator/v10"

// PaginationRequest mendefinisikan parameter standar untuk pagination
// Page: halaman saat ini (minimal 1)
// Limit: jumlah item per halaman (minimal 1, maksimal 100)
// Sort: field yang digunakan untuk sorting
// Order: urutan sorting (asc atau desc)
// Search: kata kunci untuk pencarian
type PaginationRequest struct {
	Page   int    `query:"page" validate:"min=1" default:"1"`
	Limit  int    `query:"limit" validate:"min=1,max=100" default:"10"`
	Sort   string `query:"sort" default:"id"`
	Order  string `query:"order" validate:"oneof=asc desc" default:"asc"`
	Search string `query:"search"`
}

// GetOffset menghitung offset berdasarkan page dan limit
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// Validate melakukan validasi terhadap PaginationRequest
func (p *PaginationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// IDParam mendefinisikan parameter ID dari URL path
// ID: identifier yang diambil dari path parameter
type IDParam struct {
	ID uint `params:"id" validate:"required"`
}

// Validate melakukan validasi terhadap IDParam
func (id *IDParam) Validate() error {
	validate := validator.New()
	return validate.Struct(id)
}
