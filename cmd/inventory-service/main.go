package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/handler"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/repository"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/usecase"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/config"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/database"
	pkgkafka "github.com/suleymankursatdemir/ecommerce-platform/pkg/kafka"
	pb "github.com/suleymankursatdemir/ecommerce-platform/pkg/grpc/inventorypb"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/logger"
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

	time.Sleep(2 * time.Second)

	consumer.Close()
	logger.Info("inventory service stopped gracefully")
}
