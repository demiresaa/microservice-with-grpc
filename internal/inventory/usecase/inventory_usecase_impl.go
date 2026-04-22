package usecase

import (
	"context"
	"fmt"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/repository"
)

type inventoryUseCase struct {
	repo repository.InventoryRepository
}

func NewInventoryUseCase(repo repository.InventoryRepository) InventoryUseCase {
	return &inventoryUseCase{repo: repo}
}

func (uc *inventoryUseCase) DeductStock(ctx context.Context, event *dto.InventoryEvent) error {
	if err := uc.repo.DeductStock(ctx, event.ProductID, event.Quantity); err != nil {
		return fmt.Errorf("failed to deduct stock for product %s: %w", event.ProductID, err)
	}
	return nil
}

func (uc *inventoryUseCase) CheckStock(ctx context.Context, productID string, quantity int) (bool, int, error) {
	inv, err := uc.repo.GetByProductID(ctx, productID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to check stock for product %s: %w", productID, err)
	}

	available := inv.Available()
	return available >= quantity, available, nil
}
