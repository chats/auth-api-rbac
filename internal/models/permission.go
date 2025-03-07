package models

import (
	"time"
)

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Resource    string    `gorm:"not null" json:"resource"` // เช่น "users", "roles", "articles"
	Action      string    `gorm:"not null" json:"action"`   // เช่น "read", "write", "delete"
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ตัวช่วยสำหรับการสร้าง unique key สำหรับ permission
func (p *Permission) Key() string {
	return p.Resource + ":" + p.Action
}
