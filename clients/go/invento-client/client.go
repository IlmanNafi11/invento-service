package invento

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "http://localhost:3000"
	APIVersion     = "v1"
	UserAgent      = "Invento-Go-Client/1.0"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
}

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}

	if config.HTTPClient == nil {
		jar, _ := cookiejar.New(nil)
		config.HTTPClient = &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		}
	}

	return &Client{
		baseURL:    strings.TrimSuffix(config.BaseURL, "/"),
		httpClient: config.HTTPClient,
	}
}

func (c *Client) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/api/%s/%s", c.baseURL, APIVersion, strings.TrimPrefix(endpoint, "/"))
}

func (c *Client) setAuthHeader(req *http.Request) {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) do(method, endpoint string, body, response interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.buildURL(endpoint), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if response != nil {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *Client) GetAccessToken() string {
	return c.accessToken
}

type PaginationData struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
}

func b64encode(s string) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	remaining := []byte(s)
	for len(remaining) > 0 {
		var chunk [3]byte
		copy(chunk[:], remaining)

		val := uint(chunk[0])<<16 | uint(chunk[1])<<8 | uint(chunk[2])
		for i := 0; i < 4; i++ {
			index := (val >> uint(18-6*i)) & 0x3F
			result.WriteByte(chars[index])
		}

		if len(remaining) < 3 {
			for i := len(remaining); i < 3; i++ {
				resultStr := result.String()
				resultStr = resultStr[:len(resultStr)-1] + "="
				result.Reset()
				result.WriteString(resultStr)
			}
			break
		}

		remaining = remaining[3:]
	}

	return result.String()
}
