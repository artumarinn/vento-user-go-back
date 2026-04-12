package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPaymentReceived OrderStatus = "seña_pagada"
	StatusInProduction    OrderStatus = "en_produccion"
	StatusReadyToDeliver  OrderStatus = "listo_entregar"
	StatusDelivered       OrderStatus = "entregado"
)

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	ClientID       string      `json:"client_id"`
	ClientName     string      `json:"client_name"`
	ConversationID string      `json:"conversation_id"`
	Status         OrderStatus `json:"status"`
	Total          float64     `json:"total"`
	Items          []OrderItem `json:"items"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func NewOrder(userID, clientName, conversationID string, items []OrderItem) *Order {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	return &Order{
		ID:             uuid.New().String(),
		UserID:         userID,
		ClientName:     clientName,
		ConversationID: conversationID,
		Status:         StatusPaymentReceived,
		Total:          total,
		Items:          items,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
