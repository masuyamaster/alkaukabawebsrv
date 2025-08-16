package models

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `json:"name"`
	Email      string    `gorm:"uniqueIndex" json:"email"`
	Password   string    `json:"-"`              // bcrypt (for local provider)
	Provider   string    `json:"provider"`       // "local" or "google"
	GoogleID   string    `gorm:"index" json:"google_id"`
	PictureURL string    `json:"picture_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}