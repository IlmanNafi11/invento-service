package invento

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ProfileResponse represents the profile API response
type ProfileResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    ProfileData `json:"data"`
}

// ProfileData contains user profile and permissions
type ProfileData struct {
	User        User             `json:"user"`
	Permissions []PermissionItem `json:"permissions"`
}

// PermissionItem represents a single permission entry
type PermissionItem struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// GetProfile retrieves the authenticated user's profile
func (c *Client) GetProfile() (*ProfileData, error) {
	var resp ProfileResponse
	if err := c.do("GET", "/profile", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// DownloadProjects downloads multiple projects as a ZIP file
func (c *Client) DownloadProjects(ids []uint, destPath string) error {
	body := map[string]interface{}{
		"ids": ids,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.buildURL("/project/download"), bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write download: %w", err)
	}

	return nil
}
