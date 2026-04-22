package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/handler"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/repository"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/usecase"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/config"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/database"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/logger"
)

func main() {
	logger := logger.New("payment-service")

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgresConnection(ctx, &cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected")

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	successProducer := pkgkafka.NewProducer(cfg.Kafka.Brokers, "PaymentSuccess")
	defer successProducer.Close()

	failedProducer := pkgkafka.NewProducer(cfg.Kafka.Brokers, "PaymentFailed")
	defer failedProducer.Close()

	paymentHandler := handler.NewPaymentKafkaHandler(paymentUC, successProducer, failedProducer, logger)

	consumer := pkgkafka.NewConsumer(cfg.Kafka.Brokers, "OrderCreated", "payment-service-group", logger)
	consumer.SetHandler(paymentHandler.HandleOrderCreated)

	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		if err := consumer.Consume(srvCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("consumer error", "error", err)
			}
		}
	}()

	logger.Info("payment service started, listening for OrderCreated events")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutting down", "signal", sig)

	srvCancel()

	time.Sleep(2 * time.Second)

	consumer.Close()
	logger.Info("payment service stopped gracefully")
}
