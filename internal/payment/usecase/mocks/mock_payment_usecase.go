// ============================================================================
// MOCK - PAYMENT USECASE
// ============================================================================
// PaymentUseCase interface'inin mock implementasyonu.
// ============================================================================

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/dto"
)

type MockPaymentUseCase struct {
	mock.Mock
}

func (m *MockPaymentUseCase) ProcessPayment(ctx context.Context, event *dto.PaymentEvent) (*dto.PaymentResult, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResult), args.Error(1)
}
