package constants_test

import (
	"testing"

	"invento-service/internal/constants"

	"github.com/stretchr/testify/assert"
)

func TestRBACResourceConstants(t *testing.T) {
	assert.Equal(t, "Permission", constants.ResourcePermission)
	assert.Equal(t, "Role", constants.ResourceRole)
	assert.Equal(t, "User", constants.ResourceUser)
	assert.Equal(t, "Project", constants.ResourceProject)
	assert.Equal(t, "Modul", constants.ResourceModul)
}

func TestRBACActionConstants(t *testing.T) {
	assert.Equal(t, "read", constants.ActionRead)
	assert.Equal(t, "create", constants.ActionCreate)
	assert.Equal(t, "update", constants.ActionUpdate)
	assert.Equal(t, "delete", constants.ActionDelete)
	assert.Equal(t, "download", constants.ActionDownload)
}
