// ============================================================================
// MOCK REPOSITORY'LER
// ============================================================================
//
// Mock Nedir?
//   Bir interface'in sahte (fake) implementasyonudur.
//   Gerçek veritabanına gitmek yerine, bizim kontrol ettiğimiz veriler döner.
//
// Neden Mock Kullanırız?
//   1. Unit testler DIŞ BAĞIMLILIKTAN izole olmalıdır (DB, API, dosya vs.)
//   2. Test hızı artar (DB bağlantısı beklenmez)
//   3. Her senaryoyu kolayca simüle edebiliriz (DB hatası, boş sonuç, vs.)
//
// testify/mock Nasıl Kullanılır?
//   1. struct oluştur → embed et: mock.Mock
//   2. Interface metodlarını implemente et
//   3. Her metodda: args := m.Called(ctx, param...) ile parametreleri kaydet
//   4. Return değerlerini: args.Get(0), args.Error(1) vb. ile döndür
//
// Testte Kullanım Örneği:
//   mock := new(MockOrderRepository)
//   mock.On("GetByID", ctx, "order-1").Return(&domain.Order{...}, nil)
//   // Artık GetByID("order-1") çağrıldığında yukarıdaki değerler döner
//
//   mock.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
//   // Herhangi bir *domain.Order parametresi için nil hatası döner
//
//   mock.AssertExpectations(t) // Tüm On() çağrılarının gerçekten kullanıldığını doğrula
// ============================================================================

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
)

// MockOrderRepository, OrderRepository interface'inin mock implementasyonudur.
// Testlerde gerçek PostgreSQL yerine bu sahte repository kullanılır.
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
