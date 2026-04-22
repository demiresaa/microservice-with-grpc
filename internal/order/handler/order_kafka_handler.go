package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/usecase"
)

type OrderKafkaHandler struct {
	usecase usecase.OrderUseCase
	logger  *slog.Logger
}

type paymentEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func NewOrderKafkaHandler(uc usecase.OrderUseCase, logger *slog.Logger) *OrderKafkaHandler {
	return &OrderKafkaHandler{
		usecase: uc,
		logger:  logger,
	}
}

func (h *OrderKafkaHandler) HandlePaymentSuccess(ctx context.Context, msg []byte) error {
	var event paymentEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error("failed to unmarshal PaymentSuccess event", "error", err)
		return err
	}

	h.logger.Info("payment succeeded, marking order as paid", "order_id", event.OrderID)

	if err := h.usecase.MarkAsPaid(ctx, event.OrderID); err != nil {
		h.logger.Error("failed to mark order as paid", "order_id", event.OrderID, "error", err)
		return err
	}

	return nil
}

func (h *OrderKafkaHandler) HandlePaymentFailed(ctx context.Context, msg []byte) error {
	var event paymentEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error("failed to unmarshal PaymentFailed event", "error", err)
		return err
	}

	h.logger.Info("payment failed, marking order as failed", "order_id", event.OrderID)

	if err := h.usecase.MarkAsFailed(ctx, event.OrderID); err != nil {
		h.logger.Error("failed to mark order as failed", "order_id", event.OrderID, "error", err)
		return err
	}

	return nil
}
