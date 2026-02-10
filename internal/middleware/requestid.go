package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for storing request ID
	RequestIDContextKey = "request_id"
)

// RequestID middleware generates and adds a unique request ID to each request
// It checks for existing request ID in headers and generates a new one if not present
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID is already present in headers
		requestID := c.Get(RequestIDHeader)

		// Generate new UUID if no request ID exists
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context for use in other middleware and handlers
		c.Locals(RequestIDContextKey, requestID)

		// Add request ID to response headers
		c.Set(RequestIDHeader, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from the Fiber context
func GetRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals(RequestIDContextKey); requestID != nil {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return ""
}
