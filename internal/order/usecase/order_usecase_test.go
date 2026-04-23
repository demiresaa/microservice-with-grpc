// ============================================================================
// ORDER USECASE UNIT TESTLERİ
// ============================================================================
//
// Bu dosya OrderUseCase'in unit testlerini içerir.
// Mock repository kullanarak DB bağımlılığını ortadan kaldırırız.
//
// ============================================================================
// MOCK İLE TEST YAZMA ADIMLARI
// ============================================================================
//
// Her test fonksiyonunda şu 3A (Arrange-Act-Assert) pattern'i uygulanır:
//
//   ARRANGE (Hazırlık):
//     1. Mock repository oluştur: mockRepo := new(mocks.MockOrderRepository)
//     2. Beklenen davranışı tanımla: mockRepo.On("Create", ctx, ...).Return(nil)
//     3. UseCase oluştur: uc := usecase.NewOrderUseCase(mockRepo)
//
//   ACT (Çalıştır):
//     Test edilecek metodu çağır: result, err := uc.CreateOrder(ctx, req)
//
//   ASSERT (Doğrula):
//     1. Sonucu kontrol et: assert.NotNil(t, result)
//     2. Hata durumunu kontrol et: assert.NoError(t, err)
//     3. Mock'un beklentilerinin karşılandığını doğrula: mockRepo.AssertExpectations(t)
//
// mockRepo.AssertExpectations(t) NEDEN ÖNEMLİ?
//   → mock.On("GetByID", ...) tanımlayıp ama kodun GetByID'yi hiç çağırmadığı
//     durumları yakalar. Yani "beklenen mock çağrısı gerçekleştirildi mi?" diye kontrol eder.
//
// ============================================================================
// MOCK KULLANIMINDA DİKKAT EDİLECEKLER
// ============================================================================
//
// 1. mock.AnythingOfType("*domain.Order") → Tip kontrolü yapar, değer farketmez
// 2. Belirli bir değer: mock.On("GetByID", ctx, "order-123") → Sadece bu ID için eşleşir
// 3. mock.On(...).Once() → Bu mock'un sadece 1 kez çağrılması gerektiğini belirtir
// 4. mock.On(...).Return(nil).Run(func(args) { ... }) → Return'den önce özel logic çalıştır
// 5. Sıralı mock'lar: mock.On("GetByID", ...).Return(order1, nil).Once()
//                      mock.On("GetByID", ...).Return(nil, err).Once()
//    → İlk çağrıda order1 döner, ikincide hata döner
//
// ============================================================================

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/repository/mocks"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

// ============================================================================
// CreateOrder Testleri
// ============================================================================

func TestOrderUseCase_CreateOrder_Success(t *testing.T) {
	// ARRANGE: Mock repository ve usecase oluştur
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()
	req := &dto.CreateOrderRequest{
		CustomerID: "cust-001",
		ProductID:  "prod-001",
		Quantity:   2,
		TotalPrice: 99.90,
	}

	// Mock davranışı: Create metodu çağrılırsa nil (hata yok) döndür
	// mock.AnythingOfType → parametrenin tipini kontrol eder, değerini umursamaz
	// Çünkü usecase içinde UUID ile yeni Order oluşturuluyor, tam değerini bilemeyiz
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)

	// ACT: Test edilecek metodu çağır
	order, err := uc.CreateOrder(ctx, req)

	// ASSERT: Sonuçları doğrula
	require.NoError(t, err, "geçerli siparişte hata olmamalı")

	// require ile kritik kontroller → fail olursa devam etmez (nil dereference önler)
	require.NotNil(t, order, "sipariş nil olmamalı")

	// assert ile detaylı kontroller → fail olsa bile devam eder (birden fazla hata görebiliriz)
	assert.Equal(t, "cust-001", order.CustomerID)
	assert.Equal(t, "prod-001", order.ProductID)
	assert.Equal(t, 2, order.Quantity)
	assert.Equal(t, 99.90, order.TotalPrice)
	assert.Equal(t, domain.OrderStatusPending, order.Status,
		"yeni sipariş her zaman PENDING durumunda olmalı")
	assert.NotEmpty(t, order.ID, "sipariş ID otomatik oluşturulmalı (UUID)")

	// Mock'un beklentilerinin karşılandığını doğrula
	// → mock.On("Create", ...) gerçekten çağrıldı mı?
	mockRepo.AssertExpectations(t)
}

func TestOrderUseCase_CreateOrder_ValidationError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()

	// Geçersiz sipariş: customer_id boş
	req := &dto.CreateOrderRequest{
		CustomerID: "", // BOŞ → Validate() hata vermeli
		ProductID:  "prod-001",
		Quantity:   2,
		TotalPrice: 99.90,
	}

	// DİKKAT: Validate hatası olacağı için Create metodu hiç çağrılmamalı
	// Bu yüzden mockRepo.On("Create", ...) tanımlamıyoruz
	// Eğer tanımlayıp çağrılmazsa → AssertExpectations FAIL verir

	// ACT
	order, err := uc.CreateOrder(ctx, req)

	// ASSERT
	assert.Error(t, err, "geçersiz siparişte hata dönmeli")
	assert.Nil(t, order, "hata durumunda order nil olmalı")
	assert.ErrorIs(t, err, apperrors.ErrEmptyCustomerID,
		"boş customer_id için ErrEmptyCustomerID dönmeli")

	// Create çağrılmadığını doğrula
	mockRepo.AssertNotCalled(t, "Create",
		"validate hatasında Create metodu çağrılmamalı")
}

func TestOrderUseCase_CreateOrder_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()
	req := &dto.CreateOrderRequest{
		CustomerID: "cust-001",
		ProductID:  "prod-001",
		Quantity:   2,
		TotalPrice: 99.90,
	}

	// Mock davranışı: Create metodu DB hatası döndürsün
	dbError := errors.New("connection refused")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(dbError)

	// ACT
	order, err := uc.CreateOrder(ctx, req)

	// ASSERT
	assert.Error(t, err, "repo hatası usecase'e yansımalı")
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to save order",
		"hata mesajı bağlam içermeli")

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// GetOrderByID Testleri
// ============================================================================

func TestOrderUseCase_GetOrderByID_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()
	expectedOrder := &domain.Order{
		ID:         "order-123",
		CustomerID: "cust-001",
		ProductID:  "prod-001",
		Quantity:   2,
		TotalPrice: 99.90,
		Status:     domain.OrderStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock: Belirli bir ID için belirli bir Order döndür
	mockRepo.On("GetByID", ctx, "order-123").Return(expectedOrder, nil)

	// ACT
	order, err := uc.GetOrderByID(ctx, "order-123")

	// ASSERT
	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, "order-123", order.ID)
	assert.Equal(t, "cust-001", order.CustomerID)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCase_GetOrderByID_NotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()

	// Mock: Var olmayan ID için hata döndür
	mockRepo.On("GetByID", ctx, "nonexistent").Return(nil, apperrors.ErrOrderNotFound)

	// ACT
	order, err := uc.GetOrderByID(ctx, "nonexistent")

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, order)

	mockRepo.AssertExpectations(t)
}

// ============================================================================
// MarkAsPaid Testleri
// ============================================================================

func TestOrderUseCase_MarkAsPaid_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()
	pendingOrder := &domain.Order{
		ID:     "order-123",
		Status: domain.OrderStatusPending,
	}

	// SIRALI MOCK ÇAĞRILARI:
	// 1. Önce GetByID ile order'ı getir (PENDING durumunda)
	// 2. Sonra UpdateStatus ile durumu PAID yap
	mockRepo.On("GetByID", ctx, "order-123").Return(pendingOrder, nil)
	mockRepo.On("UpdateStatus", ctx, "order-123", domain.OrderStatusPaid).Return(nil)

	// ACT
	err := uc.MarkAsPaid(ctx, "order-123")

	// ASSERT
	assert.NoError(t, err, "PENDING → PAID geçişi başarılı olmalı")
	mockRepo.AssertExpectations(t)
}

func TestOrderUseCase_MarkAsPaid_InvalidTransition(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()
	completedOrder := &domain.Order{
		ID:     "order-123",
		Status: domain.OrderStatusCompleted, // Terminal durum → PAID'e geçemez
	}

	// Sadece GetByID çağrılır → UpdateStatus'a gitmez (çünkü geçiş reddedilecek)
	mockRepo.On("GetByID", ctx, "order-123").Return(completedOrder, nil)

	// ACT
	err := uc.MarkAsPaid(ctx, "order-123")

	// ASSERT
	assert.Error(t, err, "COMPLETED → PAID geçişi hata vermeli")

	// UpdateStatus'un çağrılmadığını doğrula
	mockRepo.AssertNotCalled(t, "UpdateStatus",
		"geçersiz durum geçişinde UpdateStatus çağrılmamalı")
	mockRepo.AssertExpectations(t)
}

func TestOrderUseCase_MarkAsPaid_OrderNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)

	ctx := context.Background()
	mockRepo.On("GetByID", ctx, "nonexistent").Return(nil, apperrors.ErrOrderNotFound)

	// ACT
	err := uc.MarkAsPaid(ctx, "nonexistent")

	// ASSERT
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "UpdateStatus")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// MarkAsFailed Testleri
// ============================================================================

func TestOrderUseCase_MarkAsFailed_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)
	ctx := context.Background()

	pendingOrder := &domain.Order{ID: "order-123", Status: domain.OrderStatusPending}
	mockRepo.On("GetByID", ctx, "order-123").Return(pendingOrder, nil)
	mockRepo.On("UpdateStatus", ctx, "order-123", domain.OrderStatusFailed).Return(nil)

	err := uc.MarkAsFailed(ctx, "order-123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderUseCase_MarkAsFailed_InvalidTransition(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)
	ctx := context.Background()

	paidOrder := &domain.Order{ID: "order-123", Status: domain.OrderStatusPaid}
	mockRepo.On("GetByID", ctx, "order-123").Return(paidOrder, nil)

	err := uc.MarkAsFailed(ctx, "order-123")

	assert.Error(t, err, "PAID → FAILED geçişi hata vermeli")
	mockRepo.AssertNotCalled(t, "UpdateStatus")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// MarkAsCancelled Testleri
// ============================================================================

func TestOrderUseCase_MarkAsCancelled_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)
	ctx := context.Background()

	pendingOrder := &domain.Order{ID: "order-123", Status: domain.OrderStatusPending}
	mockRepo.On("GetByID", ctx, "order-123").Return(pendingOrder, nil)
	mockRepo.On("UpdateStatus", ctx, "order-123", domain.OrderStatusCancelled).Return(nil)

	err := uc.MarkAsCancelled(ctx, "order-123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderUseCase_MarkAsCancelled_InvalidTransition(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	uc := NewOrderUseCase(mockRepo)
	ctx := context.Background()

	completedOrder := &domain.Order{ID: "order-123", Status: domain.OrderStatusCompleted}
	mockRepo.On("GetByID", ctx, "order-123").Return(completedOrder, nil)

	err := uc.MarkAsCancelled(ctx, "order-123")

	assert.Error(t, err, "COMPLETED → CANCELLED geçişi hata vermeli")
	mockRepo.AssertNotCalled(t, "UpdateStatus")
	mockRepo.AssertExpectations(t)
}

// ============================================================================
// TABLE-DRIVEN TEST: CreateOrder Tüm Validasyon Senaryoları
// ============================================================================
// Aynı mock davranışını paylaşan birden fazla senaryoyu tek bir testte toplar.
// Bu örnekte mockRepo.Create hiçbir zaman çağrılmayacak (hepsi validate'de hata verecek).

func TestOrderUseCase_CreateOrder_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		req           *dto.CreateOrderRequest
		expectedError error
	}{
		{
			name: "boş customer_id",
			req: &dto.CreateOrderRequest{
				CustomerID: "",
				ProductID:  "prod-001",
				Quantity:   2,
				TotalPrice: 99.90,
			},
			expectedError: apperrors.ErrEmptyCustomerID,
		},
		{
			name: "boş product_id",
			req: &dto.CreateOrderRequest{
				CustomerID: "cust-001",
				ProductID:  "",
				Quantity:   2,
				TotalPrice: 99.90,
			},
			expectedError: apperrors.ErrEmptyProductID,
		},
		{
			name: "sıfır quantity",
			req: &dto.CreateOrderRequest{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   0,
				TotalPrice: 99.90,
			},
			expectedError: apperrors.ErrInvalidQuantity,
		},
		{
			name: "negatif total_price",
			req: &dto.CreateOrderRequest{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   2,
				TotalPrice: -50.0,
			},
			expectedError: apperrors.ErrInvalidTotalPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Her sub-test için yeni mock oluştur (testler birbirini etkilemesin)
			mockRepo := new(mocks.MockOrderRepository)
			uc := NewOrderUseCase(mockRepo)

			order, err := uc.CreateOrder(context.Background(), tt.req)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.ErrorIs(t, err, tt.expectedError,
				"beklenen hata: %v, gelen: %v", tt.expectedError, err)

			mockRepo.AssertNotCalled(t, "Create",
				"validasyon hatasında repo'ya yazılmamalı")
		})
	}
}
