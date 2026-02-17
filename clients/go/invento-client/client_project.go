package invento

import (
	"fmt"
	"net/url"
	"time"
)

// Project represents a project
type Project struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	NamaProject string    `json:"nama_project"`
	Kategori    string    `json:"kategori"`
	Semester    int       `json:"semester"`
	Ukuran      string    `json:"ukuran"`
	PathFile    string    `json:"path_file"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectUpdateRequest represents a project update request
type ProjectUpdateRequest struct {
	NamaProject *string `json:"nama_project,omitempty"`
	Kategori    *string `json:"kategori,omitempty"`
	Semester    *int    `json:"semester,omitempty"`
}

// ProjectListResponse represents the project list response
type ProjectListResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    ProjectListData `json:"data"`
}

type ProjectListData struct {
	Items      []ProjectListItem `json:"items"`
	Pagination PaginationData    `json:"pagination"`
}

type ProjectListItem struct {
	ID                 uint      `json:"id"`
	NamaProject        string    `json:"nama_project"`
	Kategori           string    `json:"kategori"`
	Semester           int       `json:"semester"`
	Ukuran             string    `json:"ukuran"`
	PathFile           string    `json:"path_file"`
	TerakhirDiperbarui time.Time `json:"terakhir_diperbarui"`
}

// GetProjects retrieves a list of projects
func (c *Client) GetProjects(params map[string]string) (*ProjectListData, error) {
	endpoint := "/project"
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		endpoint += "?" + values.Encode()
	}

	var resp ProjectListResponse
	if err := c.do("GET", endpoint, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetProject retrieves a single project by ID
func (c *Client) GetProject(id uint) (*Project, error) {
	var resp struct {
		Success bool    `json:"success"`
		Message string  `json:"message"`
		Data    Project `json:"data"`
	}
	if err := c.do("GET", fmt.Sprintf("/project/%d", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// UpdateProject updates a project
func (c *Client) UpdateProject(id uint, req ProjectUpdateRequest) error {
	return c.do("PATCH", fmt.Sprintf("/project/%d", id), req, nil)
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(id uint) error {
	return c.do("DELETE", fmt.Sprintf("/project/%d", id), nil, nil)
}
