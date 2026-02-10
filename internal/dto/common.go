package dto

import (
	"time"
)

// PaginationMetadata menyimpan informasi metadata untuk pagination
// Page: halaman saat ini
// Limit: jumlah item per halaman
// TotalItems: total keseluruhan item
// TotalPages: total keseluruhan halaman
// HasNext: apakah ada halaman berikutnya
// HasPrev: apakah ada halaman sebelumnya
type PaginationMetadata struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewPaginationMetadata membuat instance PaginationMetadata baru
func NewPaginationMetadata(page, limit, totalItems int) PaginationMetadata {
	totalPages := (totalItems + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationMetadata{
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ErrorDetail menyimpan detail error pada field tertentu
// Field: nama field yang memiliki error
// Message: pesan error yang menjelaskan kesalahan
// Code: kode error opsional untuk identifikasi error
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// NewErrorDetail membuat instance ErrorDetail baru
func NewErrorDetail(field, message string) ErrorDetail {
	return ErrorDetail{
		Field:   field,
		Message: message,
	}
}

// NewErrorDetailWithCode membuat instance ErrorDetail baru dengan kode error
func NewErrorDetailWithCode(field, message, code string) ErrorDetail {
	return ErrorDetail{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// TimeFormat adalah format waktu standar yang digunakan (ISO 8601)
const TimeFormat = "2006-01-02T15:04:05Z07:00"

// FormatISO8601 memformat time.Time ke format ISO 8601
func FormatISO8601(t time.Time) string {
	return t.Format(TimeFormat)
}

// NowISO8601 mengembalikan waktu saat ini dalam format ISO 8601
func NowISO8601() string {
	return time.Now().Format(TimeFormat)
}

// ParseISO8601 memparsing string ISO 8601 ke time.Time
func ParseISO8601(s string) (time.Time, error) {
	return time.Parse(TimeFormat, s)
}

// Default values untuk pagination
const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100

	// DefaultSort adalah field default untuk sorting
	DefaultSort = "id"

	// DefaultOrder adalah urutan default untuk sorting
	DefaultOrder = "asc"
)

// Validation messages dalam Bahasa Indonesia
const (
	MsgRequired           = "Field ini wajib diisi"
	MsgEmail              = "Format email tidak valid"
	MsgMin                = "Nilai minimal tidak terpenuhi"
	MsgMax                = "Nilai maksimal terlampaui"
	MsgOneOf              = "Nilai tidak sesuai dengan opsi yang tersedia"
	MsgNumeric            = "Harus berupa angka"
	MsgAlpha              = "Harus berupa huruf"
	MsgAlphanumeric       = "Harus berupa huruf dan angka"
	MsgLength             = "Panjang karakter tidak sesuai"
	MsgDate               = "Format tanggal tidak valid"
	MsgURL                = "Format URL tidak valid"
	MsgUUID               = "Format UUID tidak valid"
	MsgPositive           = "Nilai harus positif"
	MsgNonZero            = "Nilai tidak boleh nol"
	MsgUnique             = "Nilai sudah digunakan"
	MsgNotFound           = "Data tidak ditemukan"
	MsgUnauthorized       = "Tidak memiliki akses"
	MsgForbidden          = "Akses ditolak"
	MsgConflict           = "Data sudah ada"
	MsgBadRequest         = "Request tidak valid"
	MsgInternalServer     = "Terjadi kesalahan pada server"
	MsgPayloadTooLarge    = "Ukuran data melebihi batas maksimal"
	MsgTooManyRequests    = "Terlalu banyak permintaan, silakan coba lagi nanti"
	MsgValidationError    = "Validasi gagal"
	MsgInvalidCredentials = "Kredensial tidak valid"
)
