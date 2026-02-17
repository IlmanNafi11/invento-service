package invento

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TUSUploadResponse represents the TUS upload initiation response
type TUSUploadResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Data    TUSData `json:"data"`
}

// TUSData contains the TUS upload session data
type TUSData struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
}

// InitiateUpload starts a TUS resumable upload session
func (c *Client) InitiateUpload(filePath string, metadata map[string]string) (*TUSData, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	filename := filepath.Base(filePath)
	metaParts := []string{fmt.Sprintf("filename %s", b64encode(filename))}
	for k, v := range metadata {
		metaParts = append(metaParts, fmt.Sprintf("%s %s", k, b64encode(v)))
	}
	uploadMetadata := strings.Join(metaParts, ",")

	req, err := http.NewRequest("POST", c.buildURL("/modul/upload"), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", strconv.FormatInt(fileInfo.Size(), 10))
	req.Header.Set("Upload-Metadata", uploadMetadata)
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate upload: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload initiation failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var tusResp TUSUploadResponse
	if err := json.Unmarshal(respBody, &tusResp); err != nil {
		return nil, fmt.Errorf("failed to parse TUS response: %w", err)
	}

	return &tusResp.Data, nil
}

// UploadFile uploads a file using the TUS resumable upload protocol
func (c *Client) UploadFile(filePath string, chunkSize int64, onProgress func(uploaded, total int64)) error {
	tusData, err := c.InitiateUpload(filePath, nil)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	totalSize := fileInfo.Size()
	var offset int64

	for offset < totalSize {
		end := offset + chunkSize
		if end > totalSize {
			end = totalSize
		}

		if err := c.uploadChunk(tusData.UploadURL, file, offset, end-offset); err != nil {
			return fmt.Errorf("chunk upload failed at offset %d: %w", offset, err)
		}

		offset = end
		if onProgress != nil {
			onProgress(offset, totalSize)
		}
	}

	return nil
}

// uploadChunk uploads a single chunk to the TUS upload endpoint
func (c *Client) uploadChunk(uploadURL string, file *os.File, offset, length int64) error {
	chunk := make([]byte, length)
	if _, err := file.ReadAt(chunk, offset); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read chunk: %w", err)
	}

	req, err := http.NewRequest("PATCH", uploadURL, bytes.NewReader(chunk))
	if err != nil {
		return fmt.Errorf("failed to create chunk request: %w", err)
	}

	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Content-Length", strconv.FormatInt(length, 10))
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload chunk: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chunk upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UploadFormFile uploads a file using multipart form data
func (c *Client) UploadFormFile(fieldName, filePath string, fields map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return fmt.Errorf("failed to write field %s: %w", k, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.buildURL("/project/upload"), &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload form file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("form upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
