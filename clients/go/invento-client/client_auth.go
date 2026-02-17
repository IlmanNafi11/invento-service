package invento

import "time"

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    AuthData `json:"data"`
}

type AuthData struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// User represents a user
type User struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	JenisKelamin *string   `json:"jenis_kelamin,omitempty"`
	FotoProfil   *string   `json:"foto_profil,omitempty"`
	RoleID       *uint     `json:"role_id,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Login authenticates a user
func (c *Client) Login(req LoginRequest) (*AuthData, error) {
	var resp AuthResponse
	if err := c.do("POST", "/auth/login", req, &resp); err != nil {
		return nil, err
	}
	c.accessToken = resp.Data.AccessToken
	return &resp.Data, nil
}

// Register creates a new user account
func (c *Client) Register(req RegisterRequest) (*AuthData, error) {
	var resp AuthResponse
	if err := c.do("POST", "/auth/register", req, &resp); err != nil {
		return nil, err
	}
	c.accessToken = resp.Data.AccessToken
	return &resp.Data, nil
}

// RefreshToken refreshes the access token
func (c *Client) RefreshToken() (*AuthData, error) {
	var resp AuthResponse
	if err := c.do("POST", "/auth/refresh", nil, &resp); err != nil {
		return nil, err
	}
	c.accessToken = resp.Data.AccessToken
	return &resp.Data, nil
}

// Logout logs out the current user
func (c *Client) Logout() error {
	return c.do("POST", "/auth/logout", nil, nil)
}
