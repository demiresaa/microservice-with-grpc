package repository

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/domain"
)

type InventoryRepository interface {
	GetByProductID(ctx context.Context, productID string) (*domain.Inventory, error)
	DeductStock(ctx context.Context, productID string, quantity int) error
	ReserveStock(ctx context.Context, productID string, quantity int) error
}
