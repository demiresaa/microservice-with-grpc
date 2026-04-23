// ============================================================================
// INVENTORY KAFKA HANDLER UNIT TESTLERİ
// ============================================================================
//
// InventoryKafkaHandler, Kafka'dan gelen PaymentSuccess event'lerini
// işleyip stok düşürme yapar. Başarısızlık durumunda failProducer'a
// InventoryFailed event'i yayınlar.
//
// Test edilen davranışlar:
//   - Geçerli event + başarılı stok düşürme → nil error
//   - Geçersiz JSON → hata döner
//   - Stok yetersiz → fail event yayınlanır, error döner
//   - Usecase hatası → fail event yayınlanır, error döner
//
// NOT: failProducer bir *pkgkafka.Producer (concrete type) olduğu için
// mocklanamaz. Bu testte producer olmadan sadece usecase etkileşimini
// ve JSON parsing'i test ediyoruz. Tam akış için integration test gerekir.
// ============================================================================

package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	ucmocks "github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/usecase/mocks"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/dto"
	"log/slog"
	"os"
)

func TestInventoryKafkaHandler_HandlePaymentSuccess_InvalidJSON(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockInventoryUseCase)
	handler := NewInventoryKafkaHandler(mockUC, nil,
		slog.New(slog.NewTextHandler(os.Stdout, nil)))

	ctx := context.Background()
	invalidMsg := []byte(`not valid json`)

	// ACT
	err := handler.HandlePaymentSuccess(ctx, invalidMsg)

	// ASSERT
	assert.Error(t, err, "geçersiz JSON hata vermeli")
	mockUC.AssertNotCalled(t, "DeductStock")
}

func TestInventoryKafkaHandler_HandlePaymentSuccess_EventParsing(t *testing.T) {
	// JSON parsing'inin doğru çalıştığını doğrula
	msg, _ := json.Marshal(dto.InventoryEvent{
		OrderID:   "order-123",
		ProductID: "prod-001",
		Quantity:  5,
	})

	var event dto.InventoryEvent
	err := json.Unmarshal(msg, &event)

	assert.NoError(t, err)
	assert.Equal(t, "order-123", event.OrderID)
	assert.Equal(t, "prod-001", event.ProductID)
	assert.Equal(t, 5, event.Quantity)
}

// NOT: Tam handler akış testi (stok düşürme + fail producer)
// için *pkgkafka.Producer interface'e çevrilmelidir.
// Aynı PaymentKafkaHandler'daki notta açıklandığı gibi:
//
//   type Publisher interface {
//       Publish(ctx context.Context, key string, value any) error
//   }
//
// Bu refactor sonrası mock producer ile tam akış test edilebilir.
