package handler

import (
	"context"
	"log/slog"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/usecase"
	pb "github.com/suleymankursatdemir/ecommerce-platform/pkg/grpc/inventorypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryGRPCHandler struct {
	pb.UnimplementedInventoryServiceServer
	usecase usecase.InventoryUseCase
	logger  *slog.Logger
}

func NewInventoryGRPCHandler(uc usecase.InventoryUseCase, logger *slog.Logger) *InventoryGRPCHandler {
	return &InventoryGRPCHandler{
		usecase: uc,
		logger:  logger,
	}
}

func (h *InventoryGRPCHandler) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	h.logger.Info("gRPC CheckStock called", "product_id", req.ProductId, "quantity", req.Quantity)

	available, currentStock, err := h.usecase.CheckStock(ctx, req.ProductId, int(req.Quantity))
	if err != nil {
		h.logger.Error("failed to check stock", "product_id", req.ProductId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to check stock: %v", err)
	}

	h.logger.Info("stock check result", "product_id", req.ProductId, "available", available, "current_stock", currentStock)

	return &pb.CheckStockResponse{
		Available:    available,
		CurrentStock: int32(currentStock),
	}, nil
}
