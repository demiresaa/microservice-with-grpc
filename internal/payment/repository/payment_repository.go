package repository

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}
