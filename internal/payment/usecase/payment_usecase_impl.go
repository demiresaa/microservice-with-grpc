package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/repository"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type paymentUseCase struct {
	repo repository.PaymentRepository
}

func NewPaymentUseCase(repo repository.PaymentRepository) PaymentUseCase {
	return &paymentUseCase{repo: repo}
}

func (uc *paymentUseCase) ProcessPayment(ctx context.Context, event *dto.PaymentEvent) (*dto.PaymentResult, error) {
	payment := &domain.Payment{
		ID:        uuid.New().String(),
		OrderID:   event.OrderID,
		Amount:    event.Amount,
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	simulated := uc.simulateBalanceCheck(event.Amount)

	var status domain.PaymentStatus
	if simulated {
		status = domain.PaymentStatusSuccess
	} else {
		status = domain.PaymentStatusFailed
	}

	if err := uc.repo.UpdateStatus(ctx, payment.ID, string(status)); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	if !simulated {
		return &dto.PaymentResult{
			OrderID:   event.OrderID,
			ProductID: event.ProductID,
			Quantity:  event.Quantity,
			Status:    domain.PaymentStatusFailed,
		}, apperrors.ErrInsufficientBalance
	}

	return &dto.PaymentResult{
		OrderID:   event.OrderID,
		ProductID: event.ProductID,
		Quantity:  event.Quantity,
		Status:    domain.PaymentStatusSuccess,
	}, nil
}

func (uc *paymentUseCase) simulateBalanceCheck(amount float64) bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if amount > 10000 {
		return false
	}
	return r.Float32() > 0.2
}
