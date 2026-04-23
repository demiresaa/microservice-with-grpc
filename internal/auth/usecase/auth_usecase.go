package usecase

import (
	"context"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/dto"
)

type AuthUseCase interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshRequest) (*dto.RefreshResponse, error)
}
