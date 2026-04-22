package usecase

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/dto"
)

type OrderUseCase interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*domain.Order, error)
	GetOrderByID(ctx context.Context, id string) (*domain.Order, error)
	MarkAsPaid(ctx context.Context, orderID string) error
	MarkAsFailed(ctx context.Context, orderID string) error
	MarkAsCancelled(ctx context.Context, orderID string) error
}
