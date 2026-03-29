package entity

import (
	"time"
	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPendingVerification PaymentStatus = "pendiente_comprobante"
	PaymentStatusVerifying           PaymentStatus = "verificando"
	PaymentStatusVerifiedMP          PaymentStatus = "verificado_mp"
	PaymentStatusCompleted           PaymentStatus = "historial"
)

type Payment struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	ClientName     string        `json:"client_name"`
	Amount         float64       `json:"amount"`
	Status         PaymentStatus `json:"status"`
	Channel        string        `json:"channel"`
	Concept        string        `json:"concept"`
	ComprobanteURL string        `json:"comprobante_url"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func NewPayment(userID, clientName string, amount float64, concept string) *Payment {
	return &Payment{
		ID:         uuid.New().String(),
		UserID:     userID,
		ClientName: clientName,
		Amount:     amount,
		Status:     PaymentStatusPendingVerification,
		Concept:    concept,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}
