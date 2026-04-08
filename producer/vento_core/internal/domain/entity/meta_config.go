package entity

import (
	"time"
)

type MetaConfig struct {
	ID                   int64     `json:"id" db:"id"`
	UserID               string    `json:"user_id" db:"user_id"`
	PlatformID           string    `json:"platform_id" db:"platform_id"`
	Channel              string    `json:"channel" db:"channel"` // whatsapp, messenger, instagram
	WhatsAppBusinessID   *string   `json:"whatsapp_business_id" db:"whatsapp_business_id"`
	PermanentAccessToken string    `json:"permanent_access_token" db:"permanent_access_token"`
	VerifyToken          *string   `json:"verify_token" db:"verify_token"`
	AppSecret            *string   `json:"app_secret" db:"app_secret"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}
