package entity

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a label that can be attached to insumos for a user.
type Tag struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Label     string    `json:"label"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTag creates a new tag instance with a generated ID.
func NewTag(userID, label, color string) *Tag {
	return &Tag{
		ID:        uuid.New().String(),
		UserID:    userID,
		Label:     label,
		Color:     color,
		CreatedAt: time.Now().UTC(),
	}
}
