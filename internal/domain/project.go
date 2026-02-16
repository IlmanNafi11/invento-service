package domain

import "time"

type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"not null;type:uuid"`
	NamaProject string    `json:"nama_project" gorm:"not null;size:255"`
	Kategori    string    `json:"kategori" gorm:"not null;size:50"`
	Semester    int       `json:"semester" gorm:"not null"`
	Ukuran      string    `json:"ukuran" gorm:"not null;size:50"`
	PathFile    string    `json:"path_file" gorm:"not null;size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
