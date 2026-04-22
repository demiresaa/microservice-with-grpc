package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/usecase"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
)

type PaymentKafkaHandler struct {
	usecase        usecase.PaymentUseCase
	successProducer *pkgkafka.Producer
	failedProducer  *pkgkafka.Producer
	logger         *slog.Logger
}

func NewPaymentKafkaHandler(
	uc usecase.PaymentUseCase,
	successProducer *pkgkafka.Producer,
	failedProducer *pkgkafka.Producer,
	logger *slog.Logger,
) *PaymentKafkaHandler {
	return &PaymentKafkaHandler{
		usecase:         uc,
		successProducer: successProducer,
		failedProducer:  failedProducer,
		logger:          logger,
	}
}

func (h *PaymentKafkaHandler) HandleOrderCreated(ctx context.Context, msg []byte) error {
	var event dto.PaymentEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error("failed to unmarshal OrderCreated event", "error", err)
		return err
	}

	h.logger.Info("processing payment for order", "order_id", event.OrderID, "amount", event.Amount)

	result, err := h.usecase.ProcessPayment(ctx, &event)
	if err != nil {
		if errors.Is(err, apperrors.ErrInsufficientBalance) {
			h.logger.Warn("payment failed - insufficient balance", "order_id", event.OrderID)

			if pubErr := h.failedProducer.Publish(ctx, result.OrderID, result); pubErr != nil {
				h.logger.Error("failed to publish PaymentFailed event", "error", pubErr)
				return pubErr
			}
			return nil
		}
		return err
	}

	h.logger.Info("payment successful", "order_id", result.OrderID)

	if err := h.successProducer.Publish(ctx, result.OrderID, result); err != nil {
		h.logger.Error("failed to publish PaymentSuccess event", "error", err)
		return err
	}

	return nil
}
