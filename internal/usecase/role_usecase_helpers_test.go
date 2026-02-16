package usecase

import (
	goerrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "invento-service/internal/errors"
)

// assertRoleUsecaseAppError is a helper for asserting AppError in role usecase tests
func assertRoleUsecaseAppError(t *testing.T, err error, expectedCode string, expectedStatus int, expectedMessage string) {
	t.Helper()
	assert.Error(t, err)
	var appErr *apperrors.AppError
	if assert.True(t, goerrors.As(err, &appErr), "error harus bertipe *AppError") {
		assert.Equal(t, expectedCode, appErr.Code)
		assert.Equal(t, expectedStatus, appErr.HTTPStatus)
		assert.Equal(t, expectedMessage, appErr.Message)
	}
}
