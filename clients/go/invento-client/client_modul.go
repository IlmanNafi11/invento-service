package invento

import (
	"fmt"
	"net/url"
	"time"
)

// Modul represents a module
type Modul struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	FilePath  string    `json:"file_path"`
	FileType  string    `json:"file_type"`
	FileSize  int64     `json:"file_size"`
	Semester  int       `json:"semester"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ModulListResponse represents the module list response
type ModulListResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    ModulListData `json:"data"`
}

type ModulListData struct {
	Items      []ModulListItem `json:"items"`
	Pagination PaginationData  `json:"pagination"`
}

type ModulListItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	FileType  string    `json:"file_type"`
	FileSize  int64     `json:"file_size"`
	Semester  int       `json:"semester"`
	CreatedAt time.Time `json:"created_at"`
}

// ModulUpdateRequest represents a module update request
type ModulUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Semester *int    `json:"semester,omitempty"`
}

// GetModuls retrieves a list of modules
func (c *Client) GetModuls(params map[string]string) (*ModulListData, error) {
	endpoint := "/modul"
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		endpoint += "?" + values.Encode()
	}

	var resp ModulListResponse
	if err := c.do("GET", endpoint, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// UpdateModul updates a module
func (c *Client) UpdateModul(id uint, req ModulUpdateRequest) error {
	return c.do("PATCH", fmt.Sprintf("/modul/%d", id), req, nil)
}

// DeleteModul deletes a module
func (c *Client) DeleteModul(id uint) error {
	return c.do("DELETE", fmt.Sprintf("/modul/%d", id), nil, nil)
}
