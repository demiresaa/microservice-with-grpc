package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/handler"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/repository"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/usecase"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/config"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/database"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/health"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
	pb "github.com/suleymankursatdemir/ecommerce-platform/pkg/grpc/inventorypb"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/logger"
	appmiddleware "github.com/suleymankursatdemir/ecommerce-platform/pkg/middleware"
)

func main() {
	logger := logger.New("inventory-service")

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

	inventoryRepo := repository.NewPostgresInventoryRepository(db)
	inventoryUC := usecase.NewInventoryUseCase(inventoryRepo)

	failProducer := pkgkafka.NewProducer(cfg.Kafka.Brokers, "InventoryFailed")
	defer failProducer.Close()

	inventoryHandler := handler.NewInventoryKafkaHandler(inventoryUC, failProducer, logger)
	grpcHandler := handler.NewInventoryGRPCHandler(inventoryUC, logger)

	grpcServer := grpc.NewServer()
	pb.RegisterInventoryServiceServer(grpcServer, grpcHandler)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("failed to listen for gRPC", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("gRPC server starting", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	consumer := pkgkafka.NewConsumer(cfg.Kafka.Brokers, "PaymentSuccess", "inventory-service-group", logger)
	consumer.SetHandler(inventoryHandler.HandlePaymentSuccess)

	r := chi.NewRouter()
	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Recovery(logger))
	r.Use(appmiddleware.Logging(logger))
	r.Use(appmiddleware.CORS)

	healthChecker := health.NewHealthChecker(db, cfg.Kafka.Brokers[0])
	health.RegisterRoutes(r, healthChecker)

	httpPort := os.Getenv("HEALTH_PORT")
	if httpPort == "" {
		httpPort = "8083"
	}
	httpSrv := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("health check server starting", "port", httpPort)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server error", "error", err)
		}
	}()

	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()

	go func() {
		if err := consumer.Consume(srvCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Error("consumer error", "error", err)
			}
		}
	}()

	logger.Info("inventory service started, listening for PaymentSuccess events")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutting down", "signal", sig)

	srvCancel()
	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	httpSrv.Shutdown(shutdownCtx)

	time.Sleep(2 * time.Second)

	consumer.Close()
	logger.Info("inventory service stopped gracefully")
}
