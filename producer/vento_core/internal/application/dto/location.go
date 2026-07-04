package dto

import "time"

type CreateLocationRequest struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	IsDefault bool   `json:"is_default"`
}

type UpdateLocationRequest struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	IsDefault *bool  `json:"is_default"`
}

type LocationResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}
