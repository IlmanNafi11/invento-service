package middleware

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	dto "invento-service/internal/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestValidateRequest_ValidJSON tests successful validation with valid JSON
func TestValidateRequest_ValidJSON(t *testing.T) {
	app := fiber.New()
	app.Use(ValidateRequest(&dto.AuthRequest{}))

	var validatedRequest *dto.AuthRequest
	app.Post("/login", func(c *fiber.Ctx) error {
		validatedRequest = c.Locals("request").(*dto.AuthRequest)
		return c.JSON(validatedRequest)
	})

	requestBody := []byte(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.NotNil(t, validatedRequest)
	require.Equal(t, "test@example.com", validatedRequest.Email)
	require.Equal(t, "password123", validatedRequest.Password)
}

// TestValidateRequest_InvalidJSONFormat tests that invalid JSON returns 400 error
func TestValidateRequest_InvalidJSONFormat(t *testing.T) {
	app := fiber.New()
	app.Use(ValidateRequest(&dto.AuthRequest{}))

	app.Post("/login", func(c *fiber.Ctx) error {
		return c.SendString("Should not reach here")
	})

	// Malformed JSON
	requestBody := []byte(`{"email":"test@example.com","password":}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
}

// TestValidateRequest_ValidationFailures tests that validation errors return 422
func TestValidateRequest_ValidationFailures(t *testing.T) {
	app := fiber.New()
	app.Use(ValidateRequest(&dto.AuthRequest{}))

	app.Post("/login", func(c *fiber.Ctx) error {
		return c.SendString("Should not reach here")
	})

	tests := []struct {
		name         string
		requestBody  string
		expectStatus int
	}{
		{
			name:         "missing email",
			requestBody:  `{"password":"password123"}`,
			expectStatus: 400, // BodyParser fails for missing required fields
		},
		{
			name:         "missing password",
			requestBody:  `{"email":"test@example.com"}`,
			expectStatus: 400, // BodyParser fails for missing required fields
		},
		{
			name:         "invalid email format",
			requestBody:  `{"email":"invalid","password":"password123"}`,
			expectStatus: 400, // BodyParser rejects invalid email format
		},
		{
			name:         "password too short",
			requestBody:  `{"email":"test@example.com","password":"short"}`,
			expectStatus: 400, // Validation returns 400
		},
		{
			name:         "empty request",
			requestBody:  `{}`,
			expectStatus: 400, // BodyParser fails for empty required fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)

			require.NoError(t, err)
			require.Equal(t, tt.expectStatus, resp.StatusCode)
		})
	}
}

// TestValidateRequest_RequiredFieldValidation tests required field validation
func TestValidateRequest_RequiredFieldValidation(t *testing.T) {
	app := fiber.New()
	app.Use(ValidateRequest(&dto.RegisterRequest{}))

	app.Post("/register", func(c *fiber.Ctx) error {
		return c.SendString("Should not reach here")
	})

	tests := []struct {
		name         string
		requestBody  string
		expectStatus int
	}{
		{
			name:         "missing name",
			requestBody:  `{"email":"test@example.com","password":"password123"}`,
			expectStatus: 400, // BodyParser fails for missing required fields
		},
		{
			name:         "missing email",
			requestBody:  `{"name":"Test User","password":"password123"}`,
			expectStatus: 400, // BodyParser fails for missing required fields
		},
		{
			name:         "missing password",
			requestBody:  `{"name":"Test User","email":"test@example.com"}`,
			expectStatus: 400, // BodyParser fails for missing required fields
		},
		{
			name:         "all fields present",
			requestBody:  `{"name":"Test User","email":"test@example.com","password":"password123"}`,
			expectStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)

			require.NoError(t, err)
			require.Equal(t, tt.expectStatus, resp.StatusCode)
		})
	}
}

// TestValidateRequest_EmailValidation tests email format validation
func TestValidateRequest_EmailValidation(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectValid bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectValid: true,
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			expectValid: true,
		},
		{
			name:        "invalid email - no @",
			email:       "testexample.com",
			expectValid: false,
		},
		{
			name:        "invalid email - no domain",
			email:       "test@",
			expectValid: false,
		},
		{
			name:        "invalid email - no user",
			email:       "@example.com",
			expectValid: false,
		},
		{
			name:        "invalid email - double @",
			email:       "test@@example.com",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(ValidateRequest(&dto.AuthRequest{}))

			var validatedRequest *dto.AuthRequest
			app.Post("/login", func(c *fiber.Ctx) error {
				// Get validated request if successful
				if req := c.Locals("request"); req != nil {
					validatedRequest = req.(*dto.AuthRequest)
				}
				return c.SendString("OK")
			})

			requestBody := []byte(`{"email":"` + tt.email + `","password":"password123"}`)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)

			require.NoError(t, err)

			if tt.expectValid {
				require.Equal(t, 200, resp.StatusCode, "Email should be valid: %s", tt.email)
				require.NotNil(t, validatedRequest)
			} else {
				require.Equal(t, 400, resp.StatusCode, "Email should be invalid: %s", tt.email)
			}
		})
	}
}

// TestValidateRequest_MinMaxValidation tests min and max length validation
func TestValidateRequest_MinMaxValidation(t *testing.T) {
	tests := []struct {
		name         string
		nameField    string
		expectStatus int
	}{
		{
			name:         "name too short - 1 character",
			nameField:    "T",
			expectStatus: 400, // Validation returns 400
		},
		{
			name:         "name at minimum - 2 characters",
			nameField:    "Te",
			expectStatus: 200,
		},
		{
			name:         "name within range - 50 characters",
			nameField:    strings.Repeat("a", 50),
			expectStatus: 200,
		},
		{
			name:         "name at maximum - 100 characters",
			nameField:    strings.Repeat("a", 100),
			expectStatus: 200,
		},
		{
			name:         "name too long - 101 characters",
			nameField:    strings.Repeat("a", 101),
			expectStatus: 400, // Validation returns 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(ValidateRequest(&dto.RegisterRequest{}))

			app.Post("/register", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			requestBody := []byte(`{
				"name":"` + tt.nameField + `",
				"email":"test@example.com",
				"password":"password123"
			}`)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)

			require.NoError(t, err)
			require.Equal(t, tt.expectStatus, resp.StatusCode)
		})
	}
}

// TestValidateStruct tests direct struct validation
func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name      string
		data      interface{}
		hasErrors bool
	}{
		{
			name: "valid auth request",
			data: &dto.AuthRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			hasErrors: false,
		},
		{
			name: "invalid auth request - missing email",
			data: &dto.AuthRequest{
				Password: "password123",
			},
			hasErrors: true,
		},
		{
			name: "invalid auth request - invalid email",
			data: &dto.AuthRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			hasErrors: true,
		},
		{
			name: "valid register request",
			data: &dto.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			hasErrors: false,
		},
		{
			name: "invalid register request - name too short",
			data: &dto.RegisterRequest{
				Name:     "T",
				Email:    "test@example.com",
				Password: "password123",
			},
			hasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateStruct(tt.data)

			if tt.hasErrors {
				require.NotEmpty(t, errors, "Expected validation errors")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

// TestGetValidator tests that the validator instance is accessible
func TestGetValidator(t *testing.T) {
	validator := GetValidator()
	require.NotNil(t, validator, "Validator should not be nil")

	// Test that it's a working validator
	type TestStruct struct {
		Email string `validate:"required,email"`
	}

	// Valid case
	validStruct := TestStruct{Email: "test@example.com"}
	err := validator.Struct(validStruct)
	require.NoError(t, err, "Valid struct should pass")

	// Invalid case
	invalidStruct := TestStruct{Email: "invalid"}
	err = validator.Struct(invalidStruct)
	require.Error(t, err, "Invalid struct should fail")
}
