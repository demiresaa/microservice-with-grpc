// ============================================================================
// INVENTORY DOMAIN UNIT TESTLERİ
// ============================================================================
//
// Inventory domain'i test ediyoruz: Available(), Deduct(), Reserve()
// Bu metodlar saf iş mantığı içerir → dış bağımlılık yok → hızlı testler.
//
// Test stratejisi:
//   - Her metod için happy path + error path test edilir
//   - Sınır değerler (sıfır, tam eşit, bir fazla) kapsanır
//   - Hata durumunda struct'ın DEĞİŞMEDİĞİ doğrulanır (side-effect kontrolü)
// ============================================================================

package domain

import (
	"testing"

	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// TestInventory_Available, mevcut stok hesaplamasını test eder.
// Basit bir hesaplama → ama sıfır ve negatif gibi sınır durumlar önemli.
func TestInventory_Available(t *testing.T) {
	tests := []struct {
		name     string
		quantity int
		reserved int
		want     int
	}{
		{"normal: 100 stok, 30 rezerve → 70 müsait", 100, 30, 70},
		{"hiç rezerve yok: 50 stok, 0 rezerve → 50 müsait", 50, 0, 50},
		{"tümü rezerve: 100 stok, 100 rezerve → 0 müsait", 100, 100, 0},
		{"stok yok: 0 stok, 0 rezerve → 0 müsait", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Quantity: tt.quantity, Reserved: tt.reserved}
			assert.Equal(t, tt.want, inv.Available())
		})
	}
}

// TestInventory_Deduct, stok düşme işlemini test eder.
//
// ÖNEMLİ: Deduct metodunun yan etkisi (side effect) vardır.
//   - Başarılı düşme → inv.Quantity azalmalı
//   - Başarısız düşme → inv.Quantity DEĞİŞMEMELİ
//   - UpdatedAt alanı güncellenmeli (ama domain testinde time hassasiyeti gerektirir)
func TestInventory_Deduct(t *testing.T) {
	tests := []struct {
		name      string
		quantity  int  // Başlangıç stok miktarı
		reserved  int  // Rezerve edilmiş miktar
		deductQty int  // Düşülmek istenen miktar
		wantErr   error // Beklenen hata (nil = başarı)
		wantQty   int  // İşlem sonrası beklenen quantity değeri
	}{
		{
			name:      "başarılı: 100 stok, 20 rezerve, 30 düş → 70 kalır",
			quantity:  100,
			reserved:  20,
			deductQty: 30,
			wantErr:   nil,
			wantQty:   70,
		},
		{
			name:      "tümünü düş: 100 stok, 0 rezerve, 100 düş → 0 kalır",
			quantity:  100,
			reserved:  0,
			deductQty: 100,
			wantErr:   nil,
			wantQty:   0,
		},
		{
			name:      "yetersiz stok: 50 stok, 20 rezerve → 30 müsait, 40 düşemez",
			quantity:  50,
			reserved:  20,
			deductQty: 40, // 40 > 30 (available)
			wantErr:   apperrors.ErrInsufficientStock,
			wantQty:   50, // quantity DEĞİŞMEMELİ
		},
		{
			name:      "stoktan fazla: 10 stok, 0 rezerve, 20 düşemez",
			quantity:  10,
			reserved:  0,
			deductQty: 20,
			wantErr:   apperrors.ErrInsufficientStock,
			wantQty:   10, // quantity DEĞİŞMEMELİ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			inv := &Inventory{Quantity: tt.quantity, Reserved: tt.reserved}

			// ACT
			err := inv.Deduct(tt.deductQty)

			// ASSERT: Hem hata durumunu hem de quantity'nin son durumunu kontrol et
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantQty, inv.Quantity,
				"işlem sonrası quantity: beklenen=%d, gelen=%d", tt.wantQty, inv.Quantity)
		})
	}
}

// TestInventory_Reserve, stok rezervasyonunu test eder.
// Deduct'tan farkı: Quantity azalmaz, Reserved artar.
func TestInventory_Reserve(t *testing.T) {
	tests := []struct {
		name         string
		quantity     int
		reserved     int
		reserveQty   int
		wantErr      error
		wantReserved int
	}{
		{
			name:         "başarılı: 100 stok, 20 rezerve, 30 daha rezerve → 50 toplam",
			quantity:     100,
			reserved:     20,
			reserveQty:   30,
			wantErr:      nil,
			wantReserved: 50,
		},
		{
			name:         "tümünü rezerve: 100 stok, 0 rezerve, 100 rezerve → 100 toplam",
			quantity:     100,
			reserved:     0,
			reserveQty:   100,
			wantErr:      nil,
			wantReserved: 100,
		},
		{
			name:         "yetersiz: 50 stok, 20 rezerve → 30 müsait, 40 rezerve edemez",
			quantity:     50,
			reserved:     20,
			reserveQty:   40, // 40 > 30 (available)
			wantErr:      apperrors.ErrInsufficientStock,
			wantReserved: 20, // reserved DEĞİŞMEMELİ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Quantity: tt.quantity, Reserved: tt.reserved}
			err := inv.Reserve(tt.reserveQty)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantReserved, inv.Reserved,
				"işlem sonrası reserved: beklenen=%d, gelen=%d", tt.wantReserved, inv.Reserved)
		})
	}
}
