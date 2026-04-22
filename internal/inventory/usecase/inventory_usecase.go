package usecase

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/dto"
)

type InventoryUseCase interface {
	DeductStock(ctx context.Context, event *dto.InventoryEvent) error
	CheckStock(ctx context.Context, productID string, quantity int) (bool, int, error)
}
