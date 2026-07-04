package entity

import "time"

type BusinessProfile struct {
	ID               int64     `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	BusinessName     string    `json:"business_name" db:"business_name"`
	Description      string    `json:"description" db:"description"`
	Industry         string    `json:"industry" db:"industry"`
	Tone             string    `json:"tone" db:"tone"`
	Currency         string    `json:"currency" db:"currency"`
	DefaultAgentMode string    `json:"default_agent_mode" db:"default_agent_mode"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
