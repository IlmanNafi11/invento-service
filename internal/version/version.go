package version

// CurrentAPIVersion is the current version of the API
const CurrentAPIVersion = "v1"

// Version constants for API versioning
const (
	// V1 is the current stable version
	V1 = "v1"
)

// APIVersion represents an API version
type APIVersion string

// IsSupported checks if a version is supported
func (v APIVersion) IsSupported() bool {
	switch string(v) {
	case V1:
		// Add new versions here when they're introduced
		return true
	default:
		return false
	}
}

// IsDeprecated checks if a version is deprecated
func (v APIVersion) IsDeprecated() bool {
	// Currently v1 is not deprecated
	// When we introduce v2, v1 will become deprecated after the deprecation period
	return false
}

// GetSupportedVersions returns all currently supported API versions
func GetSupportedVersions() []APIVersion {
	return []APIVersion{V1}
}

// GetDeprecationWarning returns a deprecation warning message if applicable
func (v APIVersion) GetDeprecationWarning() string {
	if v.IsDeprecated() {
		return "Warning: This API version is deprecated and will be removed on or after [INSERT_DEPRECATION_DATE]. Please migrate to a newer version."
	}
	return ""
}
