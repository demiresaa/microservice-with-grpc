// ============================================================================
// ORDER KAFKA HANDLER UNIT TESTLERİ
// ============================================================================
//
// OrderKafkaHandler, Kafka'dan gelen ödeme event'lerini işleyip
// sipariş durumunu günceller.
//
// Handler testlerinde usecase mock'lanır → handler sadece usecase'i çağırır.
// Test odak noktası:
//   - JSON mesaj doğru parse ediliyor mu?
//   - Usecase doğru parametrelerle çağrılıyor mu?
//   - Usecase hatası handler'dan yukarı taşınıyor mu?
//   - Geçersiz JSON mesajı hata veriyor mu?
//
// ============================================================================

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ucmocks "github.com/suleymankursatdemir/ecommerce-platform/internal/order/usecase/mocks"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
	"log/slog"
	"os"
)

func newTestOrderKafkaHandler(uc *ucmocks.MockOrderUseCase) *OrderKafkaHandler {
	return NewOrderKafkaHandler(uc, slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

// ============================================================================
// HandlePaymentSuccess Testleri
// ============================================================================

func TestOrderKafkaHandler_HandlePaymentSuccess_Success(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := newTestOrderKafkaHandler(mockUC)

	ctx := context.Background()
	msg, _ := json.Marshal(map[string]string{
		"order_id": "order-123",
		"status":   "SUCCESS",
	})

	// Mock: MarkAsPaid başarılı
	mockUC.On("MarkAsPaid", ctx, "order-123").Return(nil)

	// ACT
	err := handler.HandlePaymentSuccess(ctx, msg)

	// ASSERT
	require.NoError(t, err)
	mockUC.AssertExpectations(t)
}

func TestOrderKafkaHandler_HandlePaymentSuccess_InvalidJSON(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := newTestOrderKafkaHandler(mockUC)

	ctx := context.Background()
	invalidMsg := []byte(`{invalid json`)

	// ACT
	err := handler.HandlePaymentSuccess(ctx, invalidMsg)

	// ASSERT
	assert.Error(t, err, "geçersiz JSON hata vermeli")
	// MarkAsPaid çağrılmamalı
	mockUC.AssertNotCalled(t, "MarkAsPaid")
}

func TestOrderKafkaHandler_HandlePaymentSuccess_UsecaseError(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := newTestOrderKafkaHandler(mockUC)

	ctx := context.Background()
	msg, _ := json.Marshal(map[string]string{
		"order_id": "order-123",
		"status":   "SUCCESS",
	})

	// Mock: MarkAsPaid hata dönsün (örn: geçersiz durum geçişi)
	mockUC.On("MarkAsPaid", ctx, "order-123").Return(apperrors.ErrInvalidOrderStatus)

	// ACT
	err := handler.HandlePaymentSuccess(ctx, msg)

	// ASSERT: Hata yukarı taşınmalı
	assert.Error(t, err)
	mockUC.AssertExpectations(t)
}

// ============================================================================
// HandlePaymentFailed Testleri
// ============================================================================

func TestOrderKafkaHandler_HandlePaymentFailed_Success(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := newTestOrderKafkaHandler(mockUC)

	ctx := context.Background()
	msg, _ := json.Marshal(map[string]string{
		"order_id": "order-456",
		"status":   "FAILED",
	})

	mockUC.On("MarkAsFailed", ctx, "order-456").Return(nil)

	// ACT
	err := handler.HandlePaymentFailed(ctx, msg)

	// ASSERT
	require.NoError(t, err)
	mockUC.AssertExpectations(t)
}

func TestOrderKafkaHandler_HandlePaymentFailed_InvalidJSON(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := newTestOrderKafkaHandler(mockUC)

	ctx := context.Background()
	invalidMsg := []byte(`not json at all`)

	// ACT
	err := handler.HandlePaymentFailed(ctx, invalidMsg)

	// ASSERT
	assert.Error(t, err)
	mockUC.AssertNotCalled(t, "MarkAsFailed")
}

func TestOrderKafkaHandler_HandlePaymentFailed_UsecaseError(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := newTestOrderKafkaHandler(mockUC)

	ctx := context.Background()
	msg, _ := json.Marshal(map[string]string{
		"order_id": "order-456",
		"status":   "FAILED",
	})

	mockUC.On("MarkAsFailed", ctx, "order-456").Return(errors.New("db error"))

	// ACT
	err := handler.HandlePaymentFailed(ctx, msg)

	// ASSERT
	assert.Error(t, err)
	mockUC.AssertExpectations(t)
}
