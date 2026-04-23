// ============================================================================
// MOCK - INVENTORY USECASE
// ============================================================================
// InventoryUseCase interface'inin mock implementasyonu.
// gRPC handler testlerinde kullanılır.
// ============================================================================

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/dto"
)

type MockInventoryUseCase struct {
	mock.Mock
}

func (m *MockInventoryUseCase) DeductStock(ctx context.Context, event *dto.InventoryEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockInventoryUseCase) CheckStock(ctx context.Context, productID string, quantity int) (bool, int, error) {
	args := m.Called(ctx, productID, quantity)
	return args.Bool(0), args.Int(1), args.Error(2)
}
