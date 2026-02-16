package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvConstants_AreNonEmpty(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, EnvDevelopment)
	assert.NotEmpty(t, EnvProduction)
}

func TestEnvConstants_HaveExpectedValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "development", EnvDevelopment)
	assert.Equal(t, "production", EnvProduction)
}

func TestEnvConstants_AreDifferent(t *testing.T) {
	t.Parallel()
	assert.NotEqual(t, EnvDevelopment, EnvProduction)
}
