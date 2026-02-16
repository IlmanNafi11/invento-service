package domain

import "time"

type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	NamaRole  string    `json:"nama_role" gorm:"uniqueIndex;not null;size:50"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Resource  string    `json:"resource" gorm:"not null;size:100"`
	Action    string    `json:"action" gorm:"not null;size:50"`
	Label     string    `json:"label" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RolePermission struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	RoleID       uint       `json:"role_id" gorm:"not null"`
	PermissionID uint       `json:"permission_id" gorm:"not null"`
	CreatedAt    time.Time  `json:"created_at"`
	Role         Role       `json:"role" gorm:"foreignKey:RoleID"`
	Permission   Permission `json:"permission" gorm:"foreignKey:PermissionID"`
}
