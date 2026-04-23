// ============================================================================
// ORDER HTTP HANDLER UNIT TESTLERİ
// ============================================================================
//
// OrderHandler'ın HTTP endpoint'lerini test eder.
// usecase ve inventoryClient mock'lanır.
//
// ============================================================================
// HTTP HANDLER TEST PATTERN
// ============================================================================
//
// HTTP handler'ları test etmek için httptest paketi kullanılır:
//
//   1. httptest.NewRequest(method, path, body) → sahte HTTP request oluştur
//   2. httptest.NewRecorder() → response'u yakalamak için recorder oluştur
//   3. Handler'ı çağır: handler.CreateOrder(recorder, request)
//   4. recorder.Code → HTTP status code
//   5. recorder.Body.String() → response body
//
// DİKKAT: Bu handler gRPC client (inventoryClient) ve Kafka producer
// bağımlılıklarına sahip. Producer bir concrete type olduğu için
// şimdilik sadece GetOrder endpoint'ini test ediyoruz (producer gerektirmez).
//
// CreateOrder endpoint'i için:
//   - inventoryClient bir gRPC interface → mock'lanabilir
//   - producer bir concrete struct → mocklanamaz
//   → CreateOrder tam testi için producer'ı interface'e çevirmek gerekir
// ============================================================================

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
	ucmocks "github.com/suleymankursatdemir/ecommerce-platform/internal/order/usecase/mocks"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

// ============================================================================
// GetOrder Testleri (producer/gRPC client gerektirmez)
// ============================================================================

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	mockUC := new(ucmocks.MockOrderUseCase)
	h := &OrderHandler{
		usecase:  mockUC,
		producer: nil,
	}

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

	mockUC.On("GetOrderByID", mock.Anything, "order-123").Return(expectedOrder, nil)

	r := chi.NewRouter()
	r.Get("/orders/{id}", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/orders/order-123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "başarılı istekte 200 dönmeli")

	var response domain.Order
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err, "response body JSON olmalı")
	assert.Equal(t, "order-123", response.ID)
	assert.Equal(t, "cust-001", response.CustomerID)
	assert.Equal(t, domain.OrderStatusPending, response.Status)

	mockUC.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	mockUC := new(ucmocks.MockOrderUseCase)
	h := &OrderHandler{
		usecase:  mockUC,
		producer: nil,
	}

	mockUC.On("GetOrderByID", mock.Anything, "nonexistent").Return(nil, apperrors.ErrOrderNotFound)

	r := chi.NewRouter()
	r.Get("/orders/{id}", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/orders/nonexistent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code,
		"bulunamayan siparişte 404 dönmeli")

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, apperrors.CodeOrderNotFound, response["code"])
}

func TestOrderHandler_GetOrder_EmptyID(t *testing.T) {
	mockUC := new(ucmocks.MockOrderUseCase)
	h := &OrderHandler{
		usecase:  mockUC,
		producer: nil,
	}

	r := chi.NewRouter()
	r.Get("/orders/{id}", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/orders/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code,
		"chi router boş segmenti eşleştirmez, 404 dönmeli")

	mockUC.AssertNotCalled(t, "GetOrderByID")
}

func TestOrderHandler_GetOrder_UsecaseError(t *testing.T) {
	mockUC := new(ucmocks.MockOrderUseCase)
	h := &OrderHandler{
		usecase:  mockUC,
		producer: nil,
	}

	mockUC.On("GetOrderByID", mock.Anything, "order-err").Return(nil, errors.New("unexpected"))

	r := chi.NewRouter()
	r.Get("/orders/{id}", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/orders/order-err", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
}

// ============================================================================
// CreateOrder Testleri (sınırlı - producer mocklanamıyor)
// ============================================================================

func TestOrderHandler_CreateOrder_InvalidRequestBody(t *testing.T) {
	// ARRANGE
	mockUC := new(ucmocks.MockOrderUseCase)
	handler := &OrderHandler{
		usecase:  mockUC,
		producer: nil,
	}

	// Geçersiz JSON body
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// ACT
	handler.CreateOrder(rec, req)

	// ASSERT
	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"geçersiz JSON body'de 400 dönmeli")

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "invalid request body", response["error"])

	mockUC.AssertNotCalled(t, "CreateOrder",
		"geçersiz body'de usecase çağrılmamalı")
}

// ============================================================================
// handleError Testleri (helper fonksiyon)
// ============================================================================

func TestHandleError_AppError(t *testing.T) {
	rec := httptest.NewRecorder()
	appErr := apperrors.ErrOrderNotFound
	handleError(rec, appErr)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, apperrors.CodeOrderNotFound, response["code"])
}

func TestHandleError_GenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleError(rec, errors.New("some random error"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response["error"])
}

// ============================================================================
// mapCodeToStatus Testleri
// ============================================================================

func TestMapCodeToStatus(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		wantStatus int
	}{
		{"ORDER_001 → 404", apperrors.CodeOrderNotFound, http.StatusNotFound},
		{"ORDER_004 → 400", apperrors.CodeEmptyCustomerID, http.StatusBadRequest},
		{"ORDER_005 → 400", apperrors.CodeEmptyProductID, http.StatusBadRequest},
		{"ORDER_006 → 400", apperrors.CodeInvalidQuantity, http.StatusBadRequest},
		{"ORDER_007 → 400", apperrors.CodeInvalidTotalPrice, http.StatusBadRequest},
		{"bilinmeyen kod → 500", "UNKNOWN_CODE", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, mapCodeToStatus(tt.code))
		})
	}
}
