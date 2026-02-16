package rbac_test

import (
	"testing"

	"invento-service/internal/rbac"

	"github.com/stretchr/testify/assert"
)

func TestRBACResourceConstants(t *testing.T) {
	assert.Equal(t, "Permission", rbac.ResourcePermission)
	assert.Equal(t, "Role", rbac.ResourceRole)
	assert.Equal(t, "User", rbac.ResourceUser)
	assert.Equal(t, "Project", rbac.ResourceProject)
	assert.Equal(t, "Modul", rbac.ResourceModul)
}

func TestRBACActionConstants(t *testing.T) {
	assert.Equal(t, "read", rbac.ActionRead)
	assert.Equal(t, "create", rbac.ActionCreate)
	assert.Equal(t, "update", rbac.ActionUpdate)
	assert.Equal(t, "delete", rbac.ActionDelete)
	assert.Equal(t, "download", rbac.ActionDownload)
}
