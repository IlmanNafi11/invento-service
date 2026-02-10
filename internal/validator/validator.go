package validator

import (
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"

	goPlaygroundValidator "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ValidatePasswordStrength validates that a password meets strength requirements:
// - At least 8 characters long
// - Contains at least one lowercase letter
// - Contains at least one uppercase letter
// - Contains at least one digit
// - Contains at least one special character
//
// Usage as struct tag: `validate:"password_strength"`
func ValidatePasswordStrength(fl goPlaygroundValidator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}

// ValidateFileType validates that a file matches one of the allowed types.
// Usage: `validate:"file_type=pdf,docx,zip"`
func ValidateFileType(fl goPlaygroundValidator.FieldLevel) bool {
	allowedTypesStr := fl.Param()
	if allowedTypesStr == "" {
		return true // No restriction
	}

	allowedTypes := strings.Split(allowedTypesStr, ",")

	// Handle *multipart.FileHeader type
	fileHeader, ok := fl.Field().Interface().(*multipart.FileHeader)
	if !ok || fileHeader == nil {
		return true // Skip validation if not a file header or nil
	}

	// Get file extension
	ext := strings.ToLower(strings.TrimPrefix(fileHeader.Filename, "."))
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	for _, allowedType := range allowedTypes {
		allowedType = strings.TrimSpace(strings.ToLower(allowedType))
		if !strings.HasPrefix(allowedType, ".") {
			allowedType = "." + allowedType
		}
		if ext == allowedType {
			return true
		}
	}

	return false
}

// ValidateFileSize validates that a file does not exceed the maximum size.
// Usage: `validate:"file_size=10485760"` (10MB in bytes)
func ValidateFileSize(fl goPlaygroundValidator.FieldLevel) bool {
	maxSizeStr := fl.Param()
	if maxSizeStr == "" {
		return true // No size restriction
	}

	var maxSize int64
	fmt.Sscanf(maxSizeStr, "%d", &maxSize)

	// Handle *multipart.FileHeader type
	fileHeader, ok := fl.Field().Interface().(*multipart.FileHeader)
	if !ok || fileHeader == nil {
		return true // Skip validation if not a file header or nil
	}

	return fileHeader.Size <= maxSize
}

// ValidateFile is a helper function to validate uploaded files directly.
// Useful for validating files outside struct validation context.
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
//	if err := validator.ValidateFile(file, []string{".pdf", ".docx"}, 10*1024*1024); err != nil {
//	    return c.Status(400).JSON(fiber.Map{"error": err.Error()})
//	}
func ValidateFile(fileHeader interface{}, allowedTypes []string, maxSizeBytes int64) error {
	fh, ok := fileHeader.(*multipart.FileHeader)
	if !ok || fh == nil {
		return fmt.Errorf("file is nil or not a FileHeader")
	}

	// Check file size
	if maxSizeBytes > 0 && fh.Size > maxSizeBytes {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSizeBytes)
	}

	// Check file type
	if len(allowedTypes) > 0 {
		ext := strings.ToLower(getFileExtension(fh.Filename))
		found := false
		for _, allowedType := range allowedTypes {
			allowedType = strings.ToLower(allowedType)
			if !strings.HasPrefix(allowedType, ".") {
				allowedType = "." + allowedType
			}
			if ext == allowedType {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("file type %s is not allowed. Allowed types: %s", ext, strings.Join(allowedTypes, ", "))
		}
	}

	return nil
}

// getFileExtension extracts the file extension from a filename.
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}

// GetAllowedFileTypes returns a list of commonly allowed file types for document uploads.
func GetAllowedFileTypes() []string {
	return []string{
		".pdf",              // PDF documents
		".doc", ".docx",     // Microsoft Word
		".xls", ".xlsx",     // Microsoft Excel
		".ppt", ".pptx",     // Microsoft PowerPoint
		".txt",              // Plain text
		".zip", ".rar",      // Archives
		".jpg", ".jpeg",     // JPEG images
		".png",              // PNG images
	}
}

// GetMaxUploadSize returns the maximum upload size for different contexts.
func GetMaxUploadSize(context string) int64 {
	switch context {
	case "avatar":
		return 2 * 1024 * 1024 // 2MB
	case "document":
		return 10 * 1024 * 1024 // 10MB
	case "video":
		return 100 * 1024 * 1024 // 100MB
	default:
		return 5 * 1024 * 1024 // 5MB default
	}
}

// ValidateMultipartForm validates a multipart form request with file uploads.
//
// Parameters:
//   - c: Fiber context
//   - maxMemory: Maximum memory for storing multipart form data (default: 32MB)
//
// Returns:
//   - error: Error if parsing fails
//
// Usage:
//
//	if err := validator.ValidateMultipartForm(c, 32<<20); err != nil {
//	    return c.Status(400).JSON(fiber.Map{"error": "invalid form data"})
//	}
func ValidateMultipartForm(c *fiber.Ctx, maxMemory int64) error {
	if _, err := c.Context().MultipartForm(); err != nil {
		return err
	}
	return nil
}

// ParseFormFile safely parses a file from multipart form with validation.
//
// Parameters:
//   - c: Fiber context
//   - fieldName: Name of the form field containing the file
//   - allowedTypes: Allowed file extensions
//   - maxSize: Maximum file size in bytes
//
// Returns:
//   - *multipart.FileHeader: The parsed file header
//   - error: Error if parsing or validation fails
//
// Usage:
//
//	file, err := validator.ParseFormFile(c, "document", []string{".pdf", ".docx"}, 10*1024*1024)
//	if err != nil {
//	    return c.Status(400).JSON(fiber.Map{"error": err.Error()})
//	}
func ParseFormFile(c *fiber.Ctx, fieldName string, allowedTypes []string, maxSize int64) (*multipart.FileHeader, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, fmt.Errorf("file '%s' is required: %w", fieldName, err)
	}

	if err := ValidateFile(file, allowedTypes, maxSize); err != nil {
		return nil, err
	}

	return file, nil
}
