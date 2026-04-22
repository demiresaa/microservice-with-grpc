package domain

import (
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusSuccess  PaymentStatus = "SUCCESS"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID        string        `json:"id"`
	OrderID   string        `json:"order_id"`
	Amount    float64       `json:"amount"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
