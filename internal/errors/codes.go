package errors

const (
	// ErrValidation indicates validation failures (HTTP 400)
	// Message: "Data validasi tidak valid"
	ErrValidation = "VALIDATION_ERROR"

	// ErrUnauthorized indicates authentication failures (HTTP 401)
	// Message: "Tidak memiliki akses"
	ErrUnauthorized = "UNAUTHORIZED_ERROR"

	// ErrForbidden indicates authorization failures (HTTP 403)
	// Message: "Akses ditolak"
	ErrForbidden = "FORBIDDEN_ERROR"

	// ErrEmailNotConfirmed indicates unconfirmed email (HTTP 403)
	// Message: "Email belum dikonfirmasi"
	ErrEmailNotConfirmed = "EMAIL_NOT_CONFIRMED"

	// ErrNotFound indicates resource not found (HTTP 404)
	// Message: "Data tidak ditemukan"
	ErrNotFound = "NOT_FOUND_ERROR"

	// ErrConflict indicates duplicate/conflicting resources (HTTP 409)
	// Message: "Data sudah ada"
	ErrConflict = "CONFLICT_ERROR"

	// ErrInternal indicates server-side errors (HTTP 500)
	// Message: "Terjadi kesalahan pada server"
	ErrInternal = "INTERNAL_ERROR"

	// ErrTusVersionMismatch indicates TUS protocol version mismatch (HTTP 412)
	// Message: "Versi TUS protocol tidak didukung"
	ErrTusVersionMismatch = "TUS_VERSION_MISMATCH"

	// ErrTusOffsetMismatch indicates TUS upload offset mismatch (HTTP 409)
	// Message: "Upload offset tidak sesuai. Diharapkan: X, diterima: Y"
	ErrTusOffsetMismatch = "TUS_OFFSET_MISMATCH"

	// ErrTusInactive indicates TUS upload is inactive (HTTP 423)
	// Message: "Upload tidak aktif"
	ErrTusInactive = "TUS_INACTIVE"

	// ErrTusAlreadyCompleted indicates TUS upload already completed (HTTP 409)
	// Message: "Upload sudah selesai"
	ErrTusAlreadyCompleted = "TUS_ALREADY_COMPLETED"

	// ErrPayloadTooLarge indicates request payload exceeds limits (HTTP 413)
	// Message: "Ukuran data melebihi batas maksimal"
	ErrPayloadTooLarge = "PAYLOAD_TOO_LARGE"
)
