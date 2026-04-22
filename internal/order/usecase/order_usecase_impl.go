package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/repository"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type orderUseCase struct {
	repo repository.OrderRepository
}

func NewOrderUseCase(repo repository.OrderRepository) OrderUseCase {
	return &orderUseCase{repo: repo}
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*domain.Order, error) {
	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: req.TotalPrice,
		Status:     domain.OrderStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return order, nil
}

func (uc *orderUseCase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (uc *orderUseCase) MarkAsPaid(ctx context.Context, orderID string) error {
	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for status update: %w", err)
	}

	if err := order.MarkAsPaid(); err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidOrderStatus, err)
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return fmt.Errorf("failed to update order status to paid: %w", err)
	}

	return nil
}

func (uc *orderUseCase) MarkAsFailed(ctx context.Context, orderID string) error {
	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for status update: %w", err)
	}

	if err := order.MarkAsFailed(); err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidOrderStatus, err)
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return fmt.Errorf("failed to update order status to failed: %w", err)
	}

	return nil
}

func (uc *orderUseCase) MarkAsCancelled(ctx context.Context, orderID string) error {
	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order for status update: %w", err)
	}

	if err := order.MarkAsCancelled(); err != nil {
		return apperrors.Wrap(apperrors.ErrInvalidOrderStatus, err)
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return fmt.Errorf("failed to update order status to cancelled: %w", err)
	}

	return nil
}
