package middleware

import (
	"errors"
	"reflect"

	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	customValidator "invento-service/internal/validator"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// validate is a singleton validator instance with custom validators registered
var validate = setupValidator()

// setupValidator initializes and configures the validator with all custom validators.
func setupValidator() *validator.Validate {
	v := validator.New()

	// Register custom validators from internal/validator package.
	// These registrations only fail if the tag name is invalid (compile-time guarantee),
	// so panicking on error is appropriate during initialization.
	mustRegister(v, "password_strength", customValidator.ValidatePasswordStrength)
	mustRegister(v, "file_type", customValidator.ValidateFileType)
	mustRegister(v, "file_size", customValidator.ValidateFileSize)
	mustRegister(v, "id_phone", customValidator.ValidateIndonesiaPhoneNumber)
	mustRegister(v, "id_mobile", customValidator.ValidateIndonesiaMobileNumber)
	mustRegister(v, "nik", customValidator.ValidateNIK)
	mustRegister(v, "npwp", customValidator.ValidateNPWP)
	mustRegister(v, "id_postal_code", customValidator.ValidateIndonesiaPostalCode)

	return v
}

// mustRegister registers a custom validator and panics on error.
func mustRegister(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		panic("failed to register validator " + tag + ": " + err.Error())
	}
}

// ValidateRequest creates a validation middleware for a specific request type.
// It parses the request body into the provided struct, validates it, and returns
// appropriate error responses if validation fails.
//
// Parameters:
//   - requestType: A pointer to the request struct type (e.g., &dto.LoginRequest{})
//
// Returns:
//   - fiber.Handler: Middleware function that handles validation
//
// Usage:
//
//	app.Post("/login", middleware.ValidateRequest(&dto.LoginRequest{}), func(c *fiber.Ctx) error {
//	    req := c.Locals("request").(*dto.LoginRequest)
//	    // Use validated request
//	    return c.JSON(req)
//	})
func ValidateRequest(requestType interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a new instance of the request type
		reqValue := reflect.New(reflect.TypeOf(requestType).Elem())
		req := reqValue.Interface()

		// Parse request body
		if err := c.BodyParser(req); err != nil {
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		// Validate the request
		if err := validate.Struct(req); err != nil {
			validationErrors := parseValidationErrors(err)
			return httputil.SendValidationErrorResponse(c, validationErrors)
		}

		// Store validated request in context
		c.Locals(LocalsKeyRequest, req)

		return c.Next()
	}
}

// ValidateStruct validates a struct and returns validation errors.
// This is a convenience function for manual validation in handlers.
//
// Parameters:
//   - data: The struct to validate
//
// Returns:
//   - []dto.ValidationError: List of validation errors (empty if valid)
//
// Usage:
//
//	func MyHandler(c *fiber.Ctx) error {
//	    var req MyRequest
//	    if err := c.BodyParser(&req); err != nil {
//	        return c.Status(400).JSON(fiber.Map{"error": "invalid format"})
//	    }
//
//	    if errors := middleware.ValidateStruct(&req); len(errors) > 0 {
//	        return c.Status(400).JSON(fiber.Map{"errors": errors})
//	    }
//
//	    // Process valid request
//	    return c.Next()
//	}
func ValidateStruct(data interface{}) []dto.ValidationError {
	var validationErrors []dto.ValidationError

	err := validate.Struct(data)
	if err != nil {
		validationErrors = parseValidationErrors(err)
	}

	return validationErrors
}

// parseValidationErrors converts validator.ValidationErrors into dto.ValidationError slice.
func parseValidationErrors(err error) []dto.ValidationError {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var errors []dto.ValidationError
		for _, err := range validationErrors {
			validationError := dto.ValidationError{
				Field:   err.Field(),
				Message: getCustomValidationMessage(err),
			}
			errors = append(errors, validationError)
		}
		return errors
	}
	// Fallback to helper for other error types
	return httputil.ValidateStruct(struct{}{})
}

// getCustomValidationMessage returns Indonesian validation messages for custom validators.
// Falls back to helper function for standard validators.
func getCustomValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "password_strength":
		return "Password harus minimal 8 karakter dengan kombinasi huruf besar, huruf kecil, angka, dan karakter khusus"
	case "file_type":
		return "Tipe file tidak diizinkan"
	case "file_size":
		return "Ukuran file melebihi batas maksimum"
	case "id_phone":
		return "Nomor telepon tidak valid (gunakan format Indonesia)"
	case "id_mobile":
		return "Nomor ponsel tidak valid (gunakan format Indonesia: 08xxxxxxxxxx)"
	case "nik":
		return "NIK tidak valid"
	case "npwp":
		return "NPWP tidak valid"
	case "id_postal_code":
		return "Kode pos tidak valid"
	default:
		// Use helper's existing message function for standard validators
		field := err.Field()
		param := err.Param()
		switch err.Tag() {
		case "required":
			return field + " wajib diisi"
		case "email":
			return "Format email tidak valid"
		case "min":
			return field + " minimal " + param + " karakter"
		case "max":
			return field + " maksimal " + param + " karakter"
		default:
			return field + " tidak valid"
		}
	}
}

// GetValidator returns the configured validator instance.
// Useful for direct validation in use cases or other layers.
//
// Usage:
//
//	v := middleware.GetValidator()
//	err := v.Struct(myData)
func GetValidator() *validator.Validate {
	return validate
}

// ValidateFile validates a file upload with type and size constraints.
// Wrapper around validator.ValidateFile for convenience in handlers.
//
// Parameters:
//   - fileHeader: The multipart file header to validate
//   - allowedTypes: List of allowed file extensions (e.g., []string{".pdf", ".docx"})
//   - maxSizeBytes: Maximum file size in bytes (0 for unlimited)
//
// Returns:
//   - error: Validation error if any, nil otherwise
//
// Usage:
//
//	file, err := c.FormFile("document")
//	if err != nil {
//	    return c.Status(400).JSON(fiber.Map{"error": "file required"})
//	}
//
//	if err := middleware.ValidateFile(file, []string{".pdf", ".docx"}, 10*1024*1024); err != nil {
//	    return c.Status(400).JSON(fiber.Map{"error": err.Error()})
//	}
func ValidateFile(fileHeader interface{}, allowedTypes []string, maxSizeBytes int64) error {
	return customValidator.ValidateFile(fileHeader, allowedTypes, maxSizeBytes)
}

// GetAllowedFileTypes returns a list of commonly allowed file types for document uploads.
func GetAllowedFileTypes() []string {
	return customValidator.GetAllowedFileTypes()
}

// GetMaxUploadSize returns the maximum upload size for different contexts.
func GetMaxUploadSize(context string) int64 {
	return customValidator.GetMaxUploadSize(context)
}
