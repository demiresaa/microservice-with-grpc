// ============================================================================
// MOCK - PAYMENT REPOSITORY
// ============================================================================
// PaymentRepository interface'inin mock implementasyonu.
// Testlerde gerçek DB yerine kullanılır.
// ============================================================================

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/domain"
)

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
