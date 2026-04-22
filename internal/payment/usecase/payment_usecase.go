package usecase

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/dto"
)

type PaymentUseCase interface {
	ProcessPayment(ctx context.Context, event *dto.PaymentEvent) (*dto.PaymentResult, error)
}
