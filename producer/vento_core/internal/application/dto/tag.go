package dto

import "time"

// CreateTagRequest is the input DTO for creating a new tag.
type CreateTagRequest struct {
	Label string `json:"label" binding:"required"`
	Color string `json:"color"`
}

// TagResponse is the output DTO for tag information.
type TagResponse struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}
