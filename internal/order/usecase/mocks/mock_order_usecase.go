// ============================================================================
// MOCK - ORDER USECASE
// ============================================================================
// OrderUseCase interface'inin mock implementasyonu.
// Handler testlerinde kullanılır → handler'ı usecase'ten izole eder.
// ============================================================================

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/dto"
)

type MockOrderUseCase struct {
	mock.Mock
}

func (m *MockOrderUseCase) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*domain.Order, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderUseCase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderUseCase) MarkAsPaid(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockOrderUseCase) MarkAsFailed(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockOrderUseCase) MarkAsCancelled(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}
