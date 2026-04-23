// ============================================================================
// INVENTORY GRPC HANDLER UNIT TESTLERİ
// ============================================================================
//
// InventoryGRPCHandler, gRPC üzerinden gelen CheckStock isteklerini işler.
// usecase mock'lanarak handler'ın sadece doğru parametrelerle usecase'i
// çağırdığını ve doğru gRPC response döndüğünü test ederiz.
//
// ============================================================================
// gRPC HANDLER TEST ETME ADIMLARI
// ============================================================================
//
// 1. Mock usecase oluştur
// 2. gRPC request objesi oluştur (protobuf tarafından üretilen struct)
// 3. Mock usecase'in dönüş değerini ayarla
// 4. Handler'ın CheckStock metodunu çağır
// 5. Dönen response ve error'ı assert et
// 6. Mock'un beklentilerinin karşılandığını doğrula
//
// gRPC testlerinde DİKKAT:
//   - Proto'dan üretilen struct'lar kullanılır (pb.CheckStockRequest, pb.CheckStockResponse)
//   - status.Errorf ile dönen hatalar: status.Code(err) ile kontrol edilir
//   - Handler'da UnimplementedInventoryServiceServer embed edilir → miras alınan metodlar da test edilmeli
// ============================================================================

package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ucmocks "github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/usecase/mocks"
	pb "github.com/suleymankursatdemir/ecommerce-platform/pkg/grpc/inventorypb"
	"log/slog"
	"os"
)

func newTestInventoryGRPCHandler(uc *ucmocks.MockInventoryUseCase) *InventoryGRPCHandler {
	return NewInventoryGRPCHandler(uc, slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

func TestInventoryGRPCHandler_CheckStock_Available(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockInventoryUseCase)
	handler := newTestInventoryGRPCHandler(mockUC)

	ctx := context.Background()
	req := &pb.CheckStockRequest{
		ProductId: "prod-001",
		Quantity:  10,
	}

	// Mock: Stok müsait → (true, 50, nil)
	mockUC.On("CheckStock", ctx, "prod-001", 10).Return(true, 50, nil)

	// ACT
	resp, err := handler.CheckStock(ctx, req)

	// ASSERT
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Available, "stok müsait olmalı")
	assert.Equal(t, int32(50), resp.CurrentStock, "mevcut stok 50 olmalı")

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCHandler_CheckStock_Insufficient(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockInventoryUseCase)
	handler := newTestInventoryGRPCHandler(mockUC)

	ctx := context.Background()
	req := &pb.CheckStockRequest{
		ProductId: "prod-001",
		Quantity:  100,
	}

	// Mock: Stok yetersiz → (false, 50, nil)
	mockUC.On("CheckStock", ctx, "prod-001", 100).Return(false, 50, nil)

	// ACT
	resp, err := handler.CheckStock(ctx, req)

	// ASSERT
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Available, "stok yetersiz olmalı")
	assert.Equal(t, int32(50), resp.CurrentStock)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCHandler_CheckStock_UsecaseError(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockInventoryUseCase)
	handler := newTestInventoryGRPCHandler(mockUC)

	ctx := context.Background()
	req := &pb.CheckStockRequest{
		ProductId: "nonexistent",
		Quantity:  10,
	}

	// Mock: Ürün bulunamadı → (false, 0, error)
	mockUC.On("CheckStock", ctx, "nonexistent", 10).Return(false, 0, assert.AnError)

	// ACT
	resp, err := handler.CheckStock(ctx, req)

	// ASSERT: gRPC handler Internal error dönmeli
	assert.Error(t, err, "usecase hatası gRPC error olarak dönmeli")
	assert.Nil(t, resp)

	mockUC.AssertExpectations(t)
}
