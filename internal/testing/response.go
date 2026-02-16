package testing

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertSuccess asserts that the response indicates success
func AssertSuccess(t *testing.T, resp *http.Response) {
	assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
		"Expected success status code, got %d", resp.StatusCode)
}

// AssertError asserts that the response indicates an error with the expected status code
func AssertError(t *testing.T, resp *http.Response, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.StatusCode,
		"Expected status code %d, got %d", expectedStatus, resp.StatusCode)
}

// AssertStatusCode asserts that the response has the expected status code
func AssertStatusCode(t *testing.T, resp *http.Response, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.StatusCode,
		"Expected status code %d, got %d", expectedStatus, resp.StatusCode)
}

// AssertJSONContentType asserts that the response has JSON content type
func AssertJSONContentType(t *testing.T, resp *http.Response) {
	contentType := resp.Header.Get("Content-Type")
	assert.True(t, contains(contentType, "application/json"),
		"Expected JSON content type, got %s", contentType)
}

// AssertJSONField asserts that a JSON field in the response body has the expected value
func AssertJSONField(t *testing.T, resp *http.Response, fieldPath string, expectedValue interface{}) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	value := getNestedField(body, fieldPath)
	assert.Equal(t, expectedValue, value,
		"Expected field %s to be %v, got %v", fieldPath, expectedValue, value)
}

// AssertJSONFieldExists asserts that a JSON field exists in the response body
func AssertJSONFieldExists(t *testing.T, resp *http.Response, fieldPath string) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	value := getNestedField(body, fieldPath)
	assert.NotNil(t, value, "Expected field %s to exist", fieldPath)
}

// AssertJSONFieldNotEmpty asserts that a JSON field exists and is not empty
func AssertJSONFieldNotEmpty(t *testing.T, resp *http.Response, fieldPath string) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	value := getNestedField(body, fieldPath)
	assert.NotNil(t, value, "Expected field %s to exist", fieldPath)
	assert.NotEmpty(t, value, "Expected field %s to not be empty", fieldPath)
}

// AssertJSONEqual asserts that the response body JSON matches the expected JSON
func AssertJSONEqual(t *testing.T, resp *http.Response, expected interface{}) {
	var actual interface{}
	err := json.NewDecoder(resp.Body).Decode(&actual)
	assert.NoError(t, err, "Failed to decode response body")

	assertJSONEqual(t, expected, actual)
}

// assertJSONEqual is a helper that recursively compares JSON structures
func assertJSONEqual(t *testing.T, expected, actual interface{}) {
	switch expected := expected.(type) {
	case map[string]interface{}:
		actualMap, ok := actual.(map[string]interface{})
		assert.True(t, ok, "Expected map, got %T", actual)
		if !ok {
			return
		}
		for key, expectedValue := range expected {
			actualValue, exists := actualMap[key]
			assert.True(t, exists, "Expected key %s to exist", key)
			if exists {
				assertJSONEqual(t, expectedValue, actualValue)
			}
		}
	case []interface{}:
		actualSlice, ok := actual.([]interface{})
		assert.True(t, ok, "Expected array, got %T", actual)
		if !ok {
			return
		}
		assert.Equal(t, len(expected), len(actualSlice), "Array length mismatch")
		for i := range expected {
			if i < len(actualSlice) {
				assertJSONEqual(t, expected[i], actualSlice[i])
			}
		}
	default:
		assert.Equal(t, expected, actual, "JSON value mismatch")
	}
}

// AssertResponseMessage asserts that the response has the expected message
func AssertResponseMessage(t *testing.T, resp *http.Response, expectedMessage string) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	message, exists := body["message"]
	assert.True(t, exists, "Expected 'message' field in response")
	assert.Equal(t, expectedMessage, message, "Message mismatch")
}

// AssertResponseCode asserts that the response code field matches the expected value
func AssertResponseCode(t *testing.T, resp *http.Response, expectedCode int) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	code, exists := body["code"]
	assert.True(t, exists, "Expected 'code' field in response")
	assert.Equal(t, float64(expectedCode), code, "Code mismatch")
}

// AssertSuccessField asserts that the status field is "success"
func AssertSuccessField(t *testing.T, resp *http.Response) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	status, exists := body["status"]
	assert.True(t, exists, "Expected 'status' field in response")
	assert.Equal(t, "success", status, "Expected status to be 'success'")
}

// AssertDataFieldExists asserts that the data field exists in the response
func AssertDataFieldExists(t *testing.T, resp *http.Response) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	_, exists := body["data"]
	assert.True(t, exists, "Expected 'data' field in response")
}

// AssertPaginationMeta asserts that pagination metadata exists and is valid
func AssertPaginationMeta(t *testing.T, resp *http.Response) {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err, "Failed to decode response body")

	meta, exists := body["meta"]
	assert.True(t, exists, "Expected 'meta' field in response")

	metaMap, ok := meta.(map[string]interface{})
	assert.True(t, ok, "Expected meta to be an object")

	// Check required pagination fields
	assert.Contains(t, metaMap, "page", "Expected 'page' in meta")
	assert.Contains(t, metaMap, "limit", "Expected 'limit' in meta")
	assert.Contains(t, metaMap, "total_pages", "Expected 'total_pages' in meta")
	assert.Contains(t, metaMap, "total_items", "Expected 'total_items' in meta")
}

// getNestedField retrieves a nested field from a map using dot notation
func getNestedField(data map[string]interface{}, fieldPath string) interface{} {
	keys := splitPath(fieldPath)
	var current interface{} = data

	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		default:
			return nil
		}
	}

	return current
}

// splitPath splits a dot-notation path into individual keys
func splitPath(path string) []string {
	var keys []string
	var current strings.Builder

	for _, r := range path {
		if r == '.' {
			if current.Len() > 0 {
				keys = append(keys, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		keys = append(keys, current.String())
	}

	return keys
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
