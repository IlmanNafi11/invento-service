package httputil

const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusPayloadTooLarge     = 413
	StatusInternalServerError = 500

	DefaultErrorMessage = "Terjadi kesalahan"
)

var StatusText = map[int]string{
	StatusOK:                  "OK",
	StatusCreated:             "Created",
	StatusNoContent:           "No Content",
	StatusBadRequest:          "Bad Request",
	StatusUnauthorized:        "Unauthorized",
	StatusForbidden:           "Forbidden",
	StatusNotFound:            "Not Found",
	StatusConflict:            "Conflict",
	StatusPayloadTooLarge:     "Payload Too Large",
	StatusInternalServerError: "Internal Server Error",
}

var DefaultMessages = map[int]string{
	StatusBadRequest:          "Request tidak valid",
	StatusUnauthorized:        "Tidak memiliki akses",
	StatusForbidden:           "Akses ditolak",
	StatusNotFound:            "Data tidak ditemukan",
	StatusConflict:            "Data sudah ada",
	StatusPayloadTooLarge:     "Ukuran data melebihi batas maksimal",
	StatusInternalServerError: "Terjadi kesalahan pada server",
}

func GetDefaultMessage(statusCode int) string {
	if message, ok := DefaultMessages[statusCode]; ok {
		return message
	}
	return DefaultErrorMessage
}
