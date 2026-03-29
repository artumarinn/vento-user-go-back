package entity

import (
	"time"
)

type MetaConfig struct {
	ID                    int64     `json:"id"`
	UserID                int64     `json:"user_id"`
	WhatsAppPhoneNumberID string    `json:"whatsapp_phone_number_id"`
	WhatsAppBusinessID    string    `json:"whatsapp_business_id"`
	PermanentAccessToken  string    `json:"permanent_access_token"`
	VerifyToken           string    `json:"verify_token"`
	AppSecret             string    `json:"app_secret"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
