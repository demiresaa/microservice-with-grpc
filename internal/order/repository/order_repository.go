package repository

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
}
