package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// MakeRequest creates and executes an HTTP request against the test app
func MakeRequest(app *fiber.App, method, path string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := app.Test(req)
	if err != nil {
		panic(fmt.Sprintf("failed to execute request: %v", err))
	}

	return &http.Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       resp.Body,
	}
}

// MakeAuthenticatedRequest creates a request with authentication token
func MakeAuthenticatedRequest(app *fiber.App, method, path string, body interface{}, userID string, email, role string) *http.Response {
	token := GenerateTestToken(userID, email, role)
	return MakeRequest(app, method, path, body, token)
}

// MakeRequestWithHeaders creates a request with custom headers
func MakeRequestWithHeaders(app *fiber.App, method, path string, body interface{}, headers map[string]string) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req)
	if err != nil {
		panic(fmt.Sprintf("failed to execute request: %v", err))
	}

	return &http.Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       resp.Body,
	}
}

// MakeRequestWithCookie creates a request with a cookie set
func MakeRequestWithCookie(app *fiber.App, method, path string, body interface{}, cookieName, cookieValue string) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("%s=%s", cookieName, cookieValue))

	resp, err := app.Test(req)
	if err != nil {
		panic(fmt.Sprintf("failed to execute request: %v", err))
	}

	return &http.Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       resp.Body,
	}
}

// MakeMultipartRequest creates a multipart form data request
func MakeMultipartRequest(app *fiber.App, method, path string, body *bytes.Buffer, contentType string, token string) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", contentType)

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := app.Test(req)
	if err != nil {
		panic(fmt.Sprintf("failed to execute request: %v", err))
	}

	return &http.Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       resp.Body,
	}
}

// ParseResponseBody parses the response body into the provided interface
func ParseResponseBody(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// GetResponseBodyString returns the response body as a string
func GetResponseBodyString(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// BuildURL builds a URL with query parameters
func BuildURL(basePath string, params map[string]string) string {
	if len(params) == 0 {
		return basePath
	}

	var queryParts []string
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
	}

	return fmt.Sprintf("%s?%s", basePath, strings.Join(queryParts, "&"))
}

// GetRequestURL builds a URL with query parameters for GET requests
func GetRequestURL(basePath string, params map[string]string) string {
	return BuildURL(basePath, params)
}
