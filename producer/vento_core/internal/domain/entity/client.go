package entity

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID            string
	UserID        string
	Name          string
	Phone         string
	SocialNetwork string
	SocialHandle  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewClient(userID, name, phone, socialNetwork, socialHandle string) *Client {
	return &Client{
		ID:            uuid.New().String(),
		UserID:        userID,
		Name:          name,
		Phone:         phone,
		SocialNetwork: socialNetwork,
		SocialHandle:  socialHandle,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
