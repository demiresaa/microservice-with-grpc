package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/suleymankursatdemir/ecommerce-platform/docs"
	pb "github.com/suleymankursatdemir/ecommerce-platform/pkg/grpc/inventorypb"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/handler"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/repository"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/usecase"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/config"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/database"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/health"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/logger"
)

// @title           E-Commerce Order Service API
// @version         1.0
// @description     Siparis yonetimi icin REST API - Kafka event-driven microservice mimarisi
// @host            localhost:8081
// @BasePath        /
func main() {
	logger := logger.New("order-service")

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgresConnection(ctx, &cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected", "host", cfg.DB.Host, "port", cfg.DB.Port)

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderUC := usecase.NewOrderUseCase(orderRepo)

	orderCreatedProducer := pkgkafka.NewProducer(cfg.Kafka.Brokers, "OrderCreated")
	defer orderCreatedProducer.Close()

	inventoryAddr := os.Getenv("INVENTORY_GRPC_ADDR")
	if inventoryAddr == "" {
		inventoryAddr = "localhost:50051"
	}
	grpcConn, err := grpc.NewClient(inventoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to inventory gRPC", "error", err)
		os.Exit(1)
	}
	defer grpcConn.Close()
	inventoryClient := pb.NewInventoryServiceClient(grpcConn)
	logger.Info("connected to inventory gRPC", "addr", inventoryAddr)

	orderHandler := handler.NewOrderHandler(orderUC, orderCreatedProducer, inventoryClient)
	orderKafkaHandler := handler.NewOrderKafkaHandler(orderUC, logger)

	paymentSuccessConsumer := pkgkafka.NewConsumer(cfg.Kafka.Brokers, "PaymentSuccess", "order-service-payment-success", logger)
	paymentSuccessConsumer.SetHandler(orderKafkaHandler.HandlePaymentSuccess)

	paymentFailedConsumer := pkgkafka.NewConsumer(cfg.Kafka.Brokers, "PaymentFailed", "order-service-payment-failed", logger)
	paymentFailedConsumer.SetHandler(orderKafkaHandler.HandlePaymentFailed)

	mux := http.NewServeMux()
	orderHandler.RegisterRoutes(mux)

	healthChecker := health.NewHealthChecker(db, cfg.Kafka.Brokers[0])
	health.RegisterRoutes(mux, healthChecker)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		if err := paymentSuccessConsumer.Consume(srvCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("payment success consumer error", "error", err)
			}
		}
	}()

	go func() {
		if err := paymentFailedConsumer.Consume(srvCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("payment failed consumer error", "error", err)
			}
		}
	}()

	go func() {
		logger.Info("order service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutting down", "signal", sig)

	srvCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	paymentSuccessConsumer.Close()
	paymentFailedConsumer.Close()

	logger.Info("order service stopped gracefully")
}
