package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/usecase"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
)

type InventoryKafkaHandler struct {
	usecase        usecase.InventoryUseCase
	failProducer   *pkgkafka.Producer
	logger         *slog.Logger
}

func NewInventoryKafkaHandler(
	uc usecase.InventoryUseCase,
	failProducer *pkgkafka.Producer,
	logger *slog.Logger,
) *InventoryKafkaHandler {
	return &InventoryKafkaHandler{
		usecase:      uc,
		failProducer: failProducer,
		logger:       logger,
	}
}

func (h *InventoryKafkaHandler) HandlePaymentSuccess(ctx context.Context, msg []byte) error {
	var event dto.InventoryEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error("failed to unmarshal PaymentSuccess event", "error", err)
		return err
	}

	h.logger.Info("deducting stock",
		"order_id", event.OrderID,
		"product_id", event.ProductID,
		"quantity", event.Quantity,
	)

	if err := h.usecase.DeductStock(ctx, &event); err != nil {
		h.logger.Error("stock deduction failed, publishing InventoryFailed",
			"order_id", event.OrderID,
			"error", err,
		)

		failEvent := map[string]any{
			"order_id":   event.OrderID,
			"product_id": event.ProductID,
			"reason":     err.Error(),
		}
		if pubErr := h.failProducer.Publish(ctx, event.OrderID, failEvent); pubErr != nil {
			h.logger.Error("failed to publish InventoryFailed event", "error", pubErr)
			return pubErr
		}

		return apperrors.Wrap(apperrors.ErrInsufficientStock, err)
	}

	h.logger.Info("stock deducted successfully",
		"order_id", event.OrderID,
		"product_id", event.ProductID,
	)
	return nil
}
