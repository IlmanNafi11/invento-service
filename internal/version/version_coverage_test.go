package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsSupported_Success tests version support checking
func TestIsSupported_Success(t *testing.T) {
	v1 := APIVersion(V1)
	assert.True(t, v1.IsSupported())

	v2 := APIVersion("v2")
	assert.False(t, v2.IsSupported())

	unsupported := APIVersion("v999")
	assert.False(t, unsupported.IsSupported())
}

// TestIsDeprecated_Success tests version deprecation checking
func TestIsDeprecated_Success(t *testing.T) {
	v1 := APIVersion(V1)
	assert.False(t, v1.IsDeprecated())

	v2 := APIVersion("v2")
	assert.False(t, v2.IsDeprecated())
}

// TestGetSupportedVersions_Success tests getting supported versions
func TestGetSupportedVersions_Success(t *testing.T) {
	versions := GetSupportedVersions()

	assert.NotEmpty(t, versions)
	assert.Len(t, versions, 1)
	assert.Equal(t, APIVersion(V1), versions[0])
}

// TestGetDeprecationWarning_Success tests deprecation warning
func TestGetDeprecationWarning_Success(t *testing.T) {
	v1 := APIVersion(V1)
	warning := v1.GetDeprecationWarning()
	assert.Empty(t, warning)

	// If we had a deprecated version, it would return a warning message
	v2 := APIVersion("v2")
	warning = v2.GetDeprecationWarning()
	assert.Empty(t, warning) // v2 is not deprecated in this implementation
}

// TestCurrentAPIVersion_Constant tests current API version constant
func TestCurrentAPIVersion_Constant(t *testing.T) {
	assert.Equal(t, "v1", CurrentAPIVersion)
}

// TestV1_Constant tests v1 version constant
func TestV1_Constant(t *testing.T) {
	assert.Equal(t, "v1", V1)
}
