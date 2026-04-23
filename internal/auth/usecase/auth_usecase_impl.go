package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/dto"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/repository"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	Role domain.Role `json:"role"`
}

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

type authUseCaseImpl struct {
	userRepo      repository.UserRepository
	jwtSecret     []byte
	refreshSecret []byte
}

func NewAuthUseCase(userRepo repository.UserRepository, jwtSecret string) AuthUseCase {
	return &authUseCaseImpl{
		userRepo:      userRepo,
		jwtSecret:     []byte(jwtSecret),
		refreshSecret: []byte(jwtSecret + "-refresh"),
	}
}

func (uc *authUseCaseImpl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	accessToken, accessExp, err := uc.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessExp.Sub(time.Now()).Seconds()),
		TokenType:    "Bearer",
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (uc *authUseCaseImpl) RefreshToken(ctx context.Context, req *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	token, err := jwt.ParseWithClaims(req.RefreshToken, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return uc.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID := claims.Subject
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	accessToken, accessExp, err := uc.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := uc.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(accessExp.Sub(time.Now()).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (uc *authUseCaseImpl) generateAccessToken(user *domain.User) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(accessTokenDuration)

	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ecommerce-platform",
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		Role: user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (uc *authUseCaseImpl) generateRefreshToken(user *domain.User) (string, error) {
	now := time.Now()

	claims := &jwt.RegisteredClaims{
		Issuer:    "ecommerce-platform-refresh",
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(uc.refreshSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseAccessToken(tokenString string, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
