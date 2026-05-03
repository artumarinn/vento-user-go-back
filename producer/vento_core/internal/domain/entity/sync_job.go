package entity

import (
	"time"
)

const (
	SyncJobStatusQueued     = "QUEUED"
	SyncJobStatusProcessing = "PROCESSING"
	SyncJobStatusSuccess    = "SUCCESS"
	SyncJobStatusFailed     = "FAILED"
)

type SyncJob struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Status    string    `json:"status" db:"status"`
	Progress  int       `json:"progress" db:"progress"`
	ErrorMsg  *string   `json:"error_msg,omitempty" db:"error_msg"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
