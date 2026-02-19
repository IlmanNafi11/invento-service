package dto

// ImportUsersRequest represents the API request for bulk user import (multipart form).
type ImportUsersRequest struct {
	DefaultRoleID int `form:"default_role_id" validate:"required"`
}

// ImportUserRow represents a single parsed Excel row (internal use, not exported in JSON).
type ImportUserRow struct {
	RowNumber    int
	Email        string
	Nama         string
	Password     string
	JenisKelamin string
	Role         string
}

// ImportReportRow represents a single row result in the import report.
type ImportReportRow struct {
	Baris    int    `json:"baris"`
	Email    string `json:"email"`
	Nama     string `json:"nama"`
	Status   string `json:"status"`
	Alasan   string `json:"alasan,omitempty"`
	Password string `json:"password,omitempty"`
}

// ImportReport represents the full import result.
type ImportReport struct {
	TotalBaris int               `json:"total_baris"`
	Berhasil   int               `json:"berhasil"`
	Dilewati   int               `json:"dilewati"`
	Detail     []ImportReportRow `json:"detail"`
}
