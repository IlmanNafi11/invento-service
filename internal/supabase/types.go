package supabase

type UserProfile struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	JenisKelamin string `json:"jenis_kelamin,omitempty"`
	FotoProfil   string `json:"foto_profil,omitempty"`
	RoleID       int    `json:"role_id,omitempty"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
