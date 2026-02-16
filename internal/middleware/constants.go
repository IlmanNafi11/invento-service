package middleware

// Context locals key constants used by auth middleware to store
// authenticated user information in Fiber's c.Locals().
const (
	LocalsKeyUserID      = "user_id"
	LocalsKeyUserEmail   = "user_email"
	LocalsKeyUserRole    = "user_role"
	LocalsKeyAccessToken = "access_token"
	LocalsKeyRequest     = "request"
)
