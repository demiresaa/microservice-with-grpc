// ============================================================================
// PAYMENT USECASE UNIT TESTLERİ
// ============================================================================
//
// PaymentUseCase.ProcessPayment() metodunu test eder.
// Mock repository ile DB bağımlılığı izole edilir.
//
// DİKKAT: ProcessPayment içinde simulateBalanceCheck() random kullanır.
//   → amount > 10000 ise HER ZAMAN false döner (deterministik)
//   → amount <= 10000 ise %80 başarı, %20 fail (non-deterministik)
//   → Bu yüzden sadece amount > 10000 senaryosunu test ediyoruz (deterministik)
//   → Production'da random seed sabitlenmeli veya dependency injection ile mock edilmeli
// ============================================================================

package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/repository/mocks"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

// TestPaymentUseCase_ProcessPayment_HighAmount_Fails, 10000 üzeri tutarlar
// için ödeme her zaman başarısız olur (simulateBalanceCheck false döner).
func TestPaymentUseCase_ProcessPayment_HighAmount_Fails(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPaymentRepository)
	uc := NewPaymentUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.PaymentEvent{
		OrderID:    "order-123",
		CustomerID: "cust-001",
		ProductID:  "prod-001",
		Quantity:   1,
		Amount:     15000.0, // > 10000 → her zaman FAIL
	}

	// Mock davranışları:
	// 1. Create çağrılır (PENDING olarak kaydedilir)
	// 2. UpdateStatus çağrılır (FAILED olarak güncellenir)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
	mockRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), "FAILED").Return(nil)

	// ACT
	result, err := uc.ProcessPayment(ctx, event)

	// ASSERT: ErrInsufficientBalance dönmeli
	assert.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrInsufficientBalance)
	require.NotNil(t, result)
	assert.Equal(t, "order-123", result.OrderID)

	mockRepo.AssertExpectations(t)
}

func TestPaymentUseCase_ProcessPayment_RepoCreateError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPaymentRepository)
	uc := NewPaymentUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.PaymentEvent{
		OrderID: "order-123",
		Amount:  50.0,
	}

	dbError := errors.New("db connection lost")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(dbError)

	// ACT
	result, err := uc.ProcessPayment(ctx, event)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create payment record")

	mockRepo.AssertExpectations(t)
}

func TestPaymentUseCase_ProcessPayment_RepoUpdateError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPaymentRepository)
	uc := NewPaymentUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.PaymentEvent{
		OrderID: "order-123",
		Amount:  15000.0, // > 10000 → FAILED olacak
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
	mockRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), "FAILED").
		Return(errors.New("update failed"))

	// ACT
	result, err := uc.ProcessPayment(ctx, event)

	// ASSERT: UpdateStatus hatası yukarı taşınmalı
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update payment status")

	mockRepo.AssertExpectations(t)
}

func TestPaymentUseCase_ProcessPayment_LowAmount_SucceedsOrFails(t *testing.T) {
	// DİKKAT: Bu test non-deterministic olabilir çünkü random kullanılıyor.
	// Sadece repo çağrılarının doğru yapıldığını test ediyoruz.
	// Production'da simulateBalanceCheck mock edilmeli veya seed sabitlenmeli.

	mockRepo := new(mocks.MockPaymentRepository)
	uc := NewPaymentUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.PaymentEvent{
		OrderID:    "order-456",
		CustomerID: "cust-001",
		ProductID:  "prod-001",
		Quantity:   1,
		Amount:     100.0, // <= 10000 → %80 başarı şansı
	}

	// Her iki senaryo için de Create çağrılır
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
	// UpdateStatus her iki durumda da çağrılır (SUCCESS veya FAILED)
	mockRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(nil)

	// ACT
	result, err := uc.ProcessPayment(ctx, event)

	// ASSERT: Sonuç her ne olursa olsun repo çağrıları yapılmış olmalı
	// (Başarı veya fail her iki durumda da Create + UpdateStatus çağrılır)
	if err != nil {
		assert.ErrorIs(t, err, apperrors.ErrInsufficientBalance)
	} else {
		require.NotNil(t, result)
		assert.Equal(t, "order-456", result.OrderID)
	}

	mockRepo.AssertExpectations(t)
}
