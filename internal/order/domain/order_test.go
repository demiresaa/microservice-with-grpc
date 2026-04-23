// ============================================================================
// ORDER DOMAIN UNIT TESTLERİ
// ============================================================================
//
// Bu dosya Order domain katmanının unit testlerini içerir.
// Domain testleri DIŞ BAĞIMLILIK (database, API, vs.) gerektirmez.
// Sadece iş mantığı test edilir → en hızlı ve en güvenilir test türüdür.
//
// ============================================================================
// GO TEST TEMELLERİ - HIZLI BAŞLANGIÇ
// ============================================================================
//
// 1) Test dosyası adı mutlaka xxx_test.go ile bitmelidir.
//    → Go derleyicisi sadece _test.go ile biten dosyalardaki test fonksiyonlarını çalıştırır.
//
// 2) Test fonksiyonlarının imzası: func TestXxx(t *testing.T)
//    → Fonksiyon adı "Test" ile başlamalıdır (büyük harf).
//    → Parametre olarak *testing.T almalıdır.
//
// 3) Çalıştırma komutları:
//    go test ./...                    → tüm paketlerdeki testleri çalıştır
//    go test ./internal/order/domain/ → sadece bu paketin testlerini çalıştır
//    go test -v ./...                 → verbose mod (her test case'in adını gösterir)
//    go test -run TestOrder_Validate  → sadece belirli bir test fonksiyonunu çalıştır
//
// 4) testify kütüphanesi:
//    → assert  : test başarısız olursa test devam eder (birden fazla hata görebilirsin)
//    → require : test başarısız olursa anında durur (kritik kontrol için)
//    Örnek: require.NotNil(t, order) → order nil ise test anında biter
//
// ============================================================================

package domain

import (
	"testing"

	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// TABLE-DRIVEN TEST PATTERN
// ============================================================================
//
// Go'nun en yaygın test stilidir. Farklı girdi/çıktı kombinasyonlarını
// bir struct slice'ı olarak tanımlayıp, tek bir döngüyle hepsini test eder.
//
// Avantajları:
//   - Yeni test case eklemek kolay (sadece slice'a yeni struct ekle)
//   - Test kodu tekrarından kurtulursun
//   - Her case'in adı (name) açıklayıcı olur → hangi case'in fail olduğu belli olur
//
// Yapı:
//
//	tests := []struct {
//	    name    string   // Test case'in açıklayıcı adı
//	    // ... girdi alanları (input)
//	    // ... beklenen çıktı alanları (expected)
//	}{
//	    { name: "başarı senaryosu", ... },
//	    { name: "hata senaryosu 1", ... },
//	}
//
//	for _, tt := range tests {
//	    t.Run(tt.name, func(t *testing.T) {
//	        // ARRANGE (hazırlık) + ACT (çalıştır) + ASSERT (doğrula)
//	    })
//	}
//
// t.Run() her case'i ayrı bir sub-test olarak çalıştırır.
// → go test -v ile her case'in sonucunu ayrı ayrı görürsün.
// → Bir case fail olursa diğerleri etkilenmez.
// ============================================================================

// TestOrder_Validate, Order.Validate() metodunun tüm senaryolarını test eder.
//
// Test edilen davranışlar:
//   - Geçerli sipariş → nil error dönmeli
//   - Boş customer_id → ErrEmptyCustomerID dönmeli
//   - Boş product_id  → ErrEmptyProductID dönmeli
//   - Sıfır/negatif quantity → ErrInvalidQuantity dönmeli
//   - Sıfır/negatif total_price → ErrInvalidTotalPrice dönmeli
func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string   // Hangi senaryo test ediliyor ( Türkçe açıklayıcı )
		order   Order    // Test edilecek Order nesnesi (ARRANGE - girdi)
		wantErr error    // Beklenen hata (ASSERT - beklenen çıktı). nil = hata beklenmiyor
	}{
		{
			name: "geçerli sipariş - tüm alanlar doğru",
			order: Order{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   2,
				TotalPrice: 99.90,
			},
			wantErr: nil, // nil → Validate() hata DÖNMEMELİ
		},
		{
			name: "boş customer_id - hata vermeli",
			order: Order{
				CustomerID: "", // Sınır değer: boş string
				ProductID:  "prod-001",
				Quantity:   2,
				TotalPrice: 99.90,
			},
			wantErr: apperrors.ErrEmptyCustomerID,
		},
		{
			name: "boş product_id - hata vermeli",
			order: Order{
				CustomerID: "cust-001",
				ProductID:  "", // Sınır değer: boş string
				Quantity:   2,
				TotalPrice: 99.90,
			},
			wantErr: apperrors.ErrEmptyProductID,
		},
		{
			name: "sıfır quantity - hata vermeli",
			order: Order{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   0, // Sınır değer: sıfır
				TotalPrice: 99.90,
			},
			wantErr: apperrors.ErrInvalidQuantity,
		},
		{
			name: "negatif quantity - hata vermeli",
			order: Order{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   -5, // Sınır değer: negatif
				TotalPrice: 99.90,
			},
			wantErr: apperrors.ErrInvalidQuantity,
		},
		{
			name: "sıfır total_price - hata vermeli",
			order: Order{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   2,
				TotalPrice: 0, // Sınır değer: sıfır
			},
			wantErr: apperrors.ErrInvalidTotalPrice,
		},
		{
			name: "negatif total_price - hata vermeli",
			order: Order{
				CustomerID: "cust-001",
				ProductID:  "prod-001",
				Quantity:   2,
				TotalPrice: -10.0, // Sınır değer: negatif
			},
			wantErr: apperrors.ErrInvalidTotalPrice,
		},
	}

	for _, tt := range tests {
		// t.Run → her test case'i ayrı bir sub-test olarak çalıştırılır
		// tt.name → çıktıda "TestOrder_Validate/geçerli_sipariş" gibi görünür
		t.Run(tt.name, func(t *testing.T) {
			// ACT: Test edilecek metodu çağır
			err := tt.order.Validate()

			// ASSERT: Sonucu doğrula
			if tt.wantErr != nil {
				// Hata bekleniyorsa → dönen hata, beklenen hata ile aynı olmalı
				// assert.Equal → pointer karşılaştırması yapar
				assert.Equal(t, tt.wantErr, err, "beklenen hata: %v, gelen: %v", tt.wantErr, err)
			} else {
				// Hata beklenmiyorsa → err nil olmalı
				assert.NoError(t, err, "geçerli siparişte hata beklenmiyordu: %v", err)
			}
		})
	}
}

// TestOrder_CanTransitionTo, sipariş durum geçişlerinin kurallarını test eder.
//
// DİKKAT: State machine testlerinde HER geçiş yönünü test etmelisin:
//   - Geçerli geçişler (true beklenenler)
//   - Geçersiz geçişler (false beklenenler)
//   - Terminal durumdan geri dönüş (completed/cancelled → hiçbir yere gidemez)
//
// Bu test PENDING → PAID gibi geçerli geçişleri VE
// PENDING → COMPLETED gibi geçersiz geçişleri kapsar.
func TestOrder_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name       string     // "from X to Y" formatında açıklayıcı
		fromStatus OrderStatus // Başlangıç durumu (ARRANGE)
		toStatus   OrderStatus // Hedef durum (ARRANGE)
		want       bool        // Geçiş izni var mı? (ASSERT)
	}{
		// --- Geçerli geçişler (true beklenir) ---
		{"PENDING → PAID başarılı", OrderStatusPending, OrderStatusPaid, true},
		{"PENDING → CANCELLED başarılı", OrderStatusPending, OrderStatusCancelled, true},
		{"PENDING → FAILED başarılı", OrderStatusPending, OrderStatusFailed, true},
		{"PAID → COMPLETED başarılı", OrderStatusPaid, OrderStatusCompleted, true},
		{"PAID → CANCELLED başarılı", OrderStatusPaid, OrderStatusCancelled, true},

		// --- Geçersiz geçişler (false beklenir) ---
		{"PENDING → COMPLETED reddedilmeli", OrderStatusPending, OrderStatusCompleted, false},
		{"PAID → PENDING reddedilmeli (geri dönüş yok)", OrderStatusPaid, OrderStatusPending, false},
		{"COMPLETED → PENDING reddedilmeli (terminal durum)", OrderStatusCompleted, OrderStatusPending, false},
		{"COMPLETED → PAID reddedilmeli (terminal durum)", OrderStatusCompleted, OrderStatusPaid, false},
		{"CANCELLED → PENDING reddedilmeli (terminal durum)", OrderStatusCancelled, OrderStatusPending, false},
		{"FAILED → PENDING reddedilmeli (terminal durum)", OrderStatusFailed, OrderStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE: Test edilecek order'ı oluştur
			o := &Order{Status: tt.fromStatus}

			// ACT + ASSERT: Geçiş kontrolü
			assert.Equal(t, tt.want, o.CanTransitionTo(tt.toStatus),
				"durum geçişi: %s → %s, beklenen: %v",
				tt.fromStatus, tt.toStatus, tt.want,
			)
		})
	}
}

// TestOrder_MarkAsPaid, MarkAsPaid metodunun durum değişimini test eder.
//
// Bu test ŞUNLARI doğrular:
//   1) Geçerli durumdan (PENDING) çağrılınca → Status PAID olmalı, error nil olmalı
//   2) Geçersiz durumdan çağrılınca → Status DEĞİŞMEMELİ, error dönmeli
//
// DİKKAT EDİLMESİ GEREKEN NOKTA:
//   Başarısız durumda Status'un DEĞİŞMEDİĞİNİ de assert etmeliyiz.
//   Sadece "error döndü" demek yetmez → status yanlışlıkla değişmiş olabilir.
func TestOrder_MarkAsPaid(t *testing.T) {
	tests := []struct {
		name    string
		status  OrderStatus // Başlangıç durumu
		wantErr bool        // Hata bekleniyor mu?
	}{
		{"PENDING → PAID başarılı", OrderStatusPending, false},
		{"COMPLETED → PAID başarısız", OrderStatusCompleted, true},
		{"PAID → PAID tekrar başarısız", OrderStatusPaid, true},
		{"CANCELLED → PAID başarısız", OrderStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			o := &Order{Status: tt.status}

			// ACT
			err := o.MarkAsPaid()

			// ASSERT
			if tt.wantErr {
				assert.Error(t, err,
					"%s durumundan PAID'e geçiş hata vermeliydi", tt.status)

				// KRİTİK: Hata durumunda status DEĞİŞMEMELİ
				assert.Equal(t, tt.status, o.Status,
					"hata durumunda status değişmemeli, eski: %s, yeni: %s",
					tt.status, o.Status)
			} else {
				assert.NoError(t, err)

				// Başarı durumunda status PAID olmalı
				assert.Equal(t, OrderStatusPaid, o.Status,
					"başarılı MarkAsPaid sonrası status PAID olmalı")
			}
		})
	}
}

// TestOrder_MarkAsCompleted, sadece PAID → COMPLETED geçişine izin verir.
// Diğer tüm durumlar hata vermeli ve status değişmemeli.
func TestOrder_MarkAsCompleted(t *testing.T) {
	tests := []struct {
		name    string
		status  OrderStatus
		wantErr bool
	}{
		{"PAID → COMPLETED başarılı", OrderStatusPaid, false},
		{"PENDING → COMPLETED başarısız", OrderStatusPending, true},
		{"COMPLETED → COMPLETED tekrar başarısız", OrderStatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			err := o.MarkAsCompleted()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, o.Status, "hata durumunda status değişmemeli")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusCompleted, o.Status)
			}
		})
	}
}

// TestOrder_MarkAsCancelled, PENDING ve PAID durumlarından CANCELLED'a geçişi test eder.
func TestOrder_MarkAsCancelled(t *testing.T) {
	tests := []struct {
		name    string
		status  OrderStatus
		wantErr bool
	}{
		{"PENDING → CANCELLED başarılı", OrderStatusPending, false},
		{"PAID → CANCELLED başarılı", OrderStatusPaid, false},
		{"COMPLETED → CANCELLED başarısız (terminal durum)", OrderStatusCompleted, true},
		{"CANCELLED → CANCELLED tekrar başarısız", OrderStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			err := o.MarkAsCancelled()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, o.Status, "hata durumunda status değişmemeli")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusCancelled, o.Status)
			}
		})
	}
}

// TestOrder_MarkAsFailed, sadece PENDING → FAILED geçişine izin verir.
func TestOrder_MarkAsFailed(t *testing.T) {
	tests := []struct {
		name    string
		status  OrderStatus
		wantErr bool
	}{
		{"PENDING → FAILED başarılı", OrderStatusPending, false},
		{"PAID → FAILED başarısız", OrderStatusPaid, true},
		{"FAILED → FAILED tekrar başarısız", OrderStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			err := o.MarkAsFailed()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, o.Status, "hata durumunda status değişmemeli")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusFailed, o.Status)
			}
		})
	}
}
