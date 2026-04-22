package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/payment/domain"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type postgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) PaymentRepository {
	return &postgresPaymentRepository{db: db}
}

func (r *postgresPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (r *postgresPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, status, created_at, updated_at
		FROM payments
		WHERE order_id = $1
	`

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotFound, err)
		}
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	return &payment, nil
}

func (r *postgresPaymentRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrPaymentNotFound
	}

	return nil
}
