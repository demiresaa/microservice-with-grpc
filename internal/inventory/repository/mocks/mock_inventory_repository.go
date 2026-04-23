// ============================================================================
// MOCK - INVENTORY REPOSITORY
// ============================================================================
// InventoryRepository interface'inin mock implementasyonu.
// Testlerde gerçek DB yerine kullanılır.
// ============================================================================

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/domain"
)

type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) GetByProductID(ctx context.Context, productID string) (*domain.Inventory, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) DeductStock(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepository) ReserveStock(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}
