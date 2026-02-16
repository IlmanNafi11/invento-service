package httputil

import (
	"errors"
	"fmt"
	"invento-service/internal/dto"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(data interface{}) []dto.ValidationError {
	var validationErrors []dto.ValidationError

	err := validate.Struct(data)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, fieldErr := range validationErrs {
				validationError := dto.ValidationError{
					Field:   fieldErr.Field(),
					Message: getValidationMessage(fieldErr),
				}
				validationErrors = append(validationErrors, validationError)
			}
		}
	}

	return validationErrors
}

func getValidationMessage(err validator.FieldError) string {
	field := err.Field()
	param := err.Param()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", field)
	case "email":
		return "Format email tidak valid"
	case "min":
		return fmt.Sprintf("%s minimal %s karakter", field, param)
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", field, param)
	case "len":
		return fmt.Sprintf("%s harus %s karakter", field, param)
	case "eq":
		return fmt.Sprintf("%s harus sama dengan %s", field, param)
	case "ne":
		return fmt.Sprintf("%s tidak boleh sama dengan %s", field, param)
	case "lt":
		return fmt.Sprintf("%s harus kurang dari %s", field, param)
	case "lte":
		return fmt.Sprintf("%s harus kurang dari atau sama dengan %s", field, param)
	case "gt":
		return fmt.Sprintf("%s harus lebih dari %s", field, param)
	case "gte":
		return fmt.Sprintf("%s harus lebih dari atau sama dengan %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s", field, param)
	case "url":
		return "Format URL tidak valid"
	case "uri":
		return "Format URI tidak valid"
	case "alpha":
		return fmt.Sprintf("%s hanya boleh berisi huruf", field)
	case "alphanum":
		return fmt.Sprintf("%s hanya boleh berisi huruf dan angka", field)
	case "numeric":
		return fmt.Sprintf("%s hanya boleh berisi angka", field)
	case "number":
		return fmt.Sprintf("%s harus berupa angka", field)
	case "hexadecimal":
		return fmt.Sprintf("%s harus berupa hexadecimal", field)
	case "hexcolor":
		return fmt.Sprintf("%s harus berupa warna hex", field)
	case "rgb":
		return fmt.Sprintf("%s harus berupa warna RGB", field)
	case "rgba":
		return fmt.Sprintf("%s harus berupa warna RGBA", field)
	case "hsl":
		return fmt.Sprintf("%s harus berupa warna HSL", field)
	case "hsla":
		return fmt.Sprintf("%s harus berupa warna HSLA", field)
	case "uuid":
		return fmt.Sprintf("%s harus berupa UUID", field)
	case "uuid3":
		return fmt.Sprintf("%s harus berupa UUID versi 3", field)
	case "uuid4":
		return fmt.Sprintf("%s harus berupa UUID versi 4", field)
	case "uuid5":
		return fmt.Sprintf("%s harus berupa UUID versi 5", field)
	case "isbn":
		return fmt.Sprintf("%s harus berupa ISBN", field)
	case "isbn10":
		return fmt.Sprintf("%s harus berupa ISBN-10", field)
	case "isbn13":
		return fmt.Sprintf("%s harus berupa ISBN-13", field)
	case "containsany":
		return fmt.Sprintf("%s harus mengandung salah satu dari: %s", field, param)
	case "contains":
		return fmt.Sprintf("%s harus mengandung: %s", field, param)
	case "excludes":
		return fmt.Sprintf("%s tidak boleh mengandung: %s", field, param)
	case "excludesall":
		return fmt.Sprintf("%s tidak boleh mengandung karakter: %s", field, param)
	case "excludesrune":
		return fmt.Sprintf("%s tidak boleh mengandung karakter: %s", field, param)
	case "startswith":
		return fmt.Sprintf("%s harus diawali dengan: %s", field, param)
	case "endswith":
		return fmt.Sprintf("%s harus diakhiri dengan: %s", field, param)
	case "datetime":
		return fmt.Sprintf("%s harus berupa tanggal dengan format: %s", field, param)
	default:
		return fmt.Sprintf("%s tidak valid", field)
	}
}
