// ============================================================================
// INVENTORY USECASE UNIT TESTLERİ
// ============================================================================
//
// InventoryUseCase'in DeductStock ve CheckStock metodlarını test eder.
// Mock repository ile DB bağımlılığı izole edilir.
//
// CheckStock: Ürün stok durumunu kontrol eder → (bool, int, error) döner
// DeductStock: Stoktan düşme yapar → error döner
//
// Her iki metod da repository'ye delege ettiği için:
//   - Repo'nun doğru parametrelerle çağrıldığını
//   - Repo hatalarının doğru şekilde yukarı taşındığını
//   test ediyoruz.
// ============================================================================

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/repository/mocks"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

// ============================================================================
// CheckStock Testleri
// ============================================================================

func TestInventoryUseCase_CheckStock_Available(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockInventoryRepository)
	uc := NewInventoryUseCase(mockRepo)

	ctx := context.Background()
	mockRepo.On("GetByProductID", ctx, "prod-001").Return(&domain.Inventory{
		ID:        "inv-1",
		ProductID: "prod-001",
		Quantity:  100,
		Reserved:  20,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	// ACT
	available, currentStock, err := uc.CheckStock(ctx, "prod-001", 50)

	// ASSERT: 100 - 20 = 80 müsait, 50 <= 80 → available = true
	require.NoError(t, err)
	assert.True(t, available, "80 müsait stoktan 50 isteniyor → müsait olmalı")
	assert.Equal(t, 80, currentStock, "mevcut stok: 100 - 20 = 80 olmalı")

	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCase_CheckStock_Insufficient(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockInventoryRepository)
	uc := NewInventoryUseCase(mockRepo)

	ctx := context.Background()
	mockRepo.On("GetByProductID", ctx, "prod-001").Return(&domain.Inventory{
		ID:        "inv-1",
		ProductID: "prod-001",
		Quantity:  50,
		Reserved:  30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	// ACT: 50 - 30 = 20 müsait, 50 isteniyor → yetersiz
	available, currentStock, err := uc.CheckStock(ctx, "prod-001", 50)

	// ASSERT
	require.NoError(t, err)
	assert.False(t, available, "20 müsait stoktan 50 isteniyor → yetersiz olmalı")
	assert.Equal(t, 20, currentStock)

	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCase_CheckStock_ProductNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockInventoryRepository)
	uc := NewInventoryUseCase(mockRepo)

	ctx := context.Background()
	mockRepo.On("GetByProductID", ctx, "nonexistent").Return(nil, apperrors.ErrProductNotFound)

	// ACT
	available, currentStock, err := uc.CheckStock(ctx, "nonexistent", 10)

	// ASSERT
	assert.Error(t, err)
	assert.False(t, available)
	assert.Equal(t, 0, currentStock)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// DeductStock Testleri
// ============================================================================

func TestInventoryUseCase_DeductStock_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockInventoryRepository)
	uc := NewInventoryUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.InventoryEvent{
		OrderID:   "order-123",
		ProductID: "prod-001",
		Quantity:  5,
	}

	mockRepo.On("DeductStock", ctx, "prod-001", 5).Return(nil)

	// ACT
	err := uc.DeductStock(ctx, event)

	// ASSERT
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCase_DeductStock_InsufficientStock(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockInventoryRepository)
	uc := NewInventoryUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.InventoryEvent{
		OrderID:   "order-123",
		ProductID: "prod-001",
		Quantity:  1000,
	}

	mockRepo.On("DeductStock", ctx, "prod-001", 1000).Return(apperrors.ErrInsufficientStock)

	// ACT
	err := uc.DeductStock(ctx, event)

	// ASSERT
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deduct stock")

	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCase_DeductStock_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockInventoryRepository)
	uc := NewInventoryUseCase(mockRepo)

	ctx := context.Background()
	event := &dto.InventoryEvent{
		OrderID:   "order-123",
		ProductID: "prod-001",
		Quantity:  5,
	}

	mockRepo.On("DeductStock", ctx, "prod-001", 5).Return(errors.New("db error"))

	// ACT
	err := uc.DeductStock(ctx, event)

	// ASSERT
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deduct stock for product prod-001")

	mockRepo.AssertExpectations(t)
}
