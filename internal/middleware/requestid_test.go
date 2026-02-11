package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestRequestID_GeneratesNewUUID tests that the middleware generates a new UUID
// when no request ID header is present
func TestRequestID_GeneratesNewUUID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	handlerCalled := false
	app.Get("/test", func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, handlerCalled)

	// Check that X-Request-ID header is present and is a valid UUID
	requestID := resp.Header.Get("X-Request-ID")
	require.NotEmpty(t, requestID, "X-Request-ID header should be present")

	parsedUUID, err := uuid.Parse(requestID)
	require.NoError(t, err, "X-Request-ID should be a valid UUID")
	require.NotEqual(t, uuid.Nil, parsedUUID, "Generated UUID should not be nil")
}

// TestRequestID_UsesExistingHeader tests that the middleware uses an existing
// request ID from the header instead of generating a new one
func TestRequestID_UsesExistingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	existingRequestID := "existing-request-id-12345"

	handlerCalled := false
	app.Get("/test", func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingRequestID)

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, handlerCalled)

	// Check that the same request ID is returned
	requestID := resp.Header.Get("X-Request-ID")
	require.Equal(t, existingRequestID, requestID, "Should use existing request ID from header")
}

// TestRequestID_ValidatesUUIDFormat tests that the generated request ID
// matches the UUID format (8-4-4-4-12 hex digits)
func TestRequestID_ValidatesUUIDFormat(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)

	requestID := resp.Header.Get("X-Request-ID")
	require.NotEmpty(t, requestID)

	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := strings.Split(requestID, "-")
	require.Len(t, parts, 5, "UUID should have 5 parts separated by hyphens")

	require.Len(t, parts[0], 8, "First part should be 8 characters")
	require.Len(t, parts[1], 4, "Second part should be 4 characters")
	require.Len(t, parts[2], 4, "Third part should be 4 characters")
	require.Len(t, parts[3], 4, "Fourth part should be 4 characters")
	require.Len(t, parts[4], 12, "Fifth part should be 12 characters")

	// All parts should be hexadecimal
	for _, part := range parts {
		for _, char := range part {
			isHex := (char >= '0' && char <= '9') ||
				(char >= 'a' && char <= 'f') ||
				(char >= 'A' && char <= 'F')
			require.True(t, isHex, "UUID should only contain hexadecimal characters")
		}
	}
}

// TestRequestID_RetrievesFromContext tests that GetRequestID correctly
// retrieves the request ID stored in the Fiber context
func TestRequestID_RetrievesFromContext(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	var retrievedRequestID string

	app.Get("/test", func(c *fiber.Ctx) error {
		retrievedRequestID = GetRequestID(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	headerRequestID := resp.Header.Get("X-Request-ID")
	require.Equal(t, headerRequestID, retrievedRequestID,
		"GetRequestID should return the same ID as in the header")
}

// TestRequestID_ContextStoresString tests that the request ID is stored
// as a string in the context locals
func TestRequestID_ContextStoresString(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	var contextValue interface{}
	var isString bool

	app.Get("/test", func(c *fiber.Ctx) error {
		contextValue = c.Locals(RequestIDContextKey)
		_, isString = contextValue.(string)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.NotNil(t, contextValue, "Request ID should be stored in context")
	require.True(t, isString, "Request ID should be stored as string type")
}

// TestRequestID_EmptyStringWhenNotSet tests that GetRequestID returns
// an empty string when the request ID middleware hasn't been applied
func TestRequestID_EmptyStringWhenNotSet(t *testing.T) {
	app := fiber.New()

	var retrievedRequestID string

	app.Get("/test", func(c *fiber.Ctx) error {
		retrievedRequestID = GetRequestID(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.Empty(t, retrievedRequestID,
		"GetRequestID should return empty string when not set")
}

// TestRequestID_PreservesThroughMiddlewareChain tests that the request ID
// is preserved through multiple middleware in the chain
func TestRequestID_PreservesThroughMiddlewareChain(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	firstMiddlewareID := ""
	secondMiddlewareID := ""

	app.Use(func(c *fiber.Ctx) error {
		firstMiddlewareID = GetRequestID(c)
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		secondMiddlewareID = GetRequestID(c)
		return c.Next()
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.NotEmpty(t, firstMiddlewareID)
	require.NotEmpty(t, secondMiddlewareID)
	require.Equal(t, firstMiddlewareID, secondMiddlewareID,
		"Request ID should be the same through the middleware chain")
}

// TestRequestID_DifferentForEachRequest tests that each request gets
// a unique request ID
func TestRequestID_DifferentForEachRequest(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Make multiple requests
	requestIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		requestIDs[i] = resp.Header.Get("X-Request-ID")
		require.NotEmpty(t, requestIDs[i])
	}

	// Check that all request IDs are unique
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		uniqueIDs[id] = true
	}

	require.Len(t, uniqueIDs, 10, "All request IDs should be unique")
}

// TestRequestID_MultipleEndpoints tests that request ID middleware
// works correctly for multiple endpoints
func TestRequestID_MultipleEndpoints(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	app.Get("/endpoint1", func(c *fiber.Ctx) error {
		return c.SendString("Endpoint 1")
	})

	app.Post("/endpoint2", func(c *fiber.Ctx) error {
		return c.SendString("Endpoint 2")
	})

	app.Put("/endpoint3", func(c *fiber.Ctx) error {
		return c.SendString("Endpoint 3")
	})

	// Test GET endpoint
	req1 := httptest.NewRequest("GET", "/endpoint1", nil)
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp1.StatusCode)
	id1 := resp1.Header.Get("X-Request-ID")
	require.NotEmpty(t, id1)

	// Test POST endpoint
	req2 := httptest.NewRequest("POST", "/endpoint2", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp2.StatusCode)
	id2 := resp2.Header.Get("X-Request-ID")
	require.NotEmpty(t, id2)

	// Test PUT endpoint
	req3 := httptest.NewRequest("PUT", "/endpoint3", nil)
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp3.StatusCode)
	id3 := resp3.Header.Get("X-Request-ID")
	require.NotEmpty(t, id3)

	// All IDs should be unique
	require.NotEqual(t, id1, id2)
	require.NotEqual(t, id2, id3)
	require.NotEqual(t, id1, id3)
}

// TestRequestID_WithExistingUUID tests that the middleware accepts
// a valid UUID format from the header
func TestRequestID_WithExistingUUID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	existingUUID := uuid.New().String()

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingUUID)

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	requestID := resp.Header.Get("X-Request-ID")
	require.Equal(t, existingUUID, requestID,
		"Should use the existing UUID from header")
}

// TestRequestID_WithCustomHeader tests that the middleware can work
// with custom header values (not just UUIDs)
func TestRequestID_WithCustomHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	customID := "my-custom-request-id-12345"

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", customID)

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	requestID := resp.Header.Get("X-Request-ID")
	require.Equal(t, customID, requestID,
		"Should use custom request ID from header")
}

// TestRequestID_NilContext tests that GetRequestID handles nil
// context values gracefully
func TestRequestID_NilContext(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	app.Use(func(c *fiber.Ctx) error {
		// Explicitly set nil to test edge case
		c.Locals(RequestIDContextKey, nil)
		return c.Next()
	})

	var retrievedRequestID string

	app.Get("/test", func(c *fiber.Ctx) error {
		retrievedRequestID = GetRequestID(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	require.Empty(t, retrievedRequestID,
		"GetRequestID should return empty string for nil context value")
}

// TestRequestID_ConstantKeys tests that the constant keys are
// correctly defined
func TestRequestID_ConstantKeys(t *testing.T) {
	require.Equal(t, "X-Request-ID", RequestIDHeader)
	require.Equal(t, "request_id", RequestIDContextKey)
}
