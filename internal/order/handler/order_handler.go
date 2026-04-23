package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/usecase"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
	pb "github.com/suleymankursatdemir/ecommerce-platform/pkg/grpc/inventorypb"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
)

type OrderHandler struct {
	usecase         usecase.OrderUseCase
	producer        *pkgkafka.Producer
	inventoryClient pb.InventoryServiceClient
}

func NewOrderHandler(uc usecase.OrderUseCase, producer *pkgkafka.Producer, inventoryClient pb.InventoryServiceClient) *OrderHandler {
	return &OrderHandler{
		usecase:         uc,
		producer:        producer,
		inventoryClient: inventoryClient,
	}
}

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	r.Post("/orders", h.CreateOrder)
	r.Get("/orders/{id}", h.GetOrder)
}

// CreateOrder godoc
// @Summary      Yeni siparis olustur
// @Description  Yeni bir siparis olusturur ve Kafka'ya OrderCreated eventi yayinlar
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateOrderRequest true "Siparis bilgileri"
// @Success      201 {object} dto.OrderResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	ctx := r.Context()

	stockResp, err := h.inventoryClient.CheckStock(ctx, &pb.CheckStockRequest{
		ProductId: req.ProductID,
		Quantity:  int32(req.Quantity),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to check stock: %v", err),
		})
		return
	}
	if !stockResp.Available {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":         "insufficient stock",
			"current_stock": fmt.Sprintf("%d", stockResp.CurrentStock),
		})
		return
	}

	order, err := h.usecase.CreateOrder(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	event := &dto.OrderEvent{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		ProductID:  order.ProductID,
		Quantity:   order.Quantity,
		TotalPrice: order.TotalPrice,
	}

	if err := h.producer.Publish(ctx, order.ID, event); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

// GetOrder godoc
// @Summary      Siparis sorgula
// @Description  ID ile siparis bilgilerini getirir
// @Tags         orders
// @Produce      json
// @Param        id path string true "Siparis ID"
// @Success      200 {object} dto.OrderResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "order id is required",
		})
		return
	}

	ctx := r.Context()

	order, err := h.usecase.GetOrderByID(ctx, id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func handleError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		status := mapCodeToStatus(appErr.Code)
		writeJSON(w, status, map[string]string{
			"code":    appErr.Code,
			"message": appErr.Message,
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "internal server error",
	})
}

func mapCodeToStatus(code string) int {
	switch {
	case len(code) >= 3 && code[:3] == "ORD" && code == "ORDER_001":
		return http.StatusNotFound
	case code == "ORDER_004" || code == "ORDER_005" || code == "ORDER_006" || code == "ORDER_007":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
